package ecs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/cloud/awsconfig"
	"github.com/jaxxstorm/landlord/internal/compute"
)

// Provider implements the compute.Provider interface for AWS ECS.
type Provider struct {
	mu               sync.RWMutex
	logger           *zap.Logger
	loadAWSConfig    func(ctx context.Context, opts awsconfig.Options) (aws.Config, error)
	tenantConfigs    map[string]*ComputeConfig
	defaultConfig    map[string]interface{}
	defaultConfigRaw json.RawMessage
}

// New creates a new ECS provider.
func New(logger *zap.Logger, defaults map[string]interface{}) *Provider {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Provider{
		logger:           logger.With(zap.String("component", "ecs-provider")),
		loadAWSConfig:    awsconfig.Load,
		tenantConfigs:    make(map[string]*ComputeConfig),
		defaultConfig:    copyConfigMap(defaults),
		defaultConfigRaw: marshalConfigMap(defaults),
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "ecs"
}

// Provision creates a new ECS service for a tenant.
func (p *Provider) Provision(ctx context.Context, spec *compute.TenantComputeSpec) (*compute.ProvisionResult, error) {
	if err := p.Validate(ctx, spec); err != nil {
		return nil, err
	}

	cfg, err := parseComputeConfig(spec.ProviderConfig, p.defaultConfig)
	if err != nil {
		return nil, err
	}

	region := resolveRegion(cfg)
	if region == "" {
		return nil, fmt.Errorf("region is required for ECS provider")
	}

	awsCfg, err := p.loadAWSConfig(ctx, awsconfig.Options{
		Region:     region,
		AssumeRole: toAssumeRoleOptions(cfg.AssumeRole),
	})
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := ecs.NewFromConfig(awsCfg)
	serviceName := resolveServiceName(cfg, spec.TenantID)

	service, err := p.describeService(ctx, client, cfg.ClusterARN, serviceName)
	if err != nil {
		return nil, err
	}

	if service == nil {
		if err := p.createService(ctx, client, cfg, spec, serviceName); err != nil {
			return nil, err
		}
	} else {
		if _, err := p.updateService(ctx, client, cfg, serviceName, service); err != nil {
			return nil, err
		}
	}

	p.storeConfig(spec.TenantID, cfg)

	return &compute.ProvisionResult{
		TenantID:      spec.TenantID,
		ProviderType:  p.Name(),
		Status:        compute.ProvisionStatusSuccess,
		ResourceIDs:   map[string]string{"service": serviceName, "cluster": cfg.ClusterARN, "task_definition": cfg.TaskDefinition},
		Message:       "ECS service provisioned",
		ProvisionedAt: time.Now(),
	}, nil
}

// Update modifies an existing ECS service for a tenant.
func (p *Provider) Update(ctx context.Context, tenantID string, spec *compute.TenantComputeSpec) (*compute.UpdateResult, error) {
	if err := p.Validate(ctx, spec); err != nil {
		return nil, err
	}

	cfg, err := parseComputeConfig(spec.ProviderConfig, p.defaultConfig)
	if err != nil {
		return nil, err
	}

	region := resolveRegion(cfg)
	if region == "" {
		return nil, fmt.Errorf("region is required for ECS provider")
	}

	awsCfg, err := p.loadAWSConfig(ctx, awsconfig.Options{
		Region:     region,
		AssumeRole: toAssumeRoleOptions(cfg.AssumeRole),
	})
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := ecs.NewFromConfig(awsCfg)
	serviceName := resolveServiceName(cfg, tenantID)

	service, err := p.describeService(ctx, client, cfg.ClusterARN, serviceName)
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	status, err := p.updateService(ctx, client, cfg, serviceName, service)
	if err != nil {
		return nil, err
	}

	p.storeConfig(tenantID, cfg)

	return &compute.UpdateResult{
		TenantID:     tenantID,
		ProviderType: p.Name(),
		Status:       status,
		Changes:      nil,
		Message:      "ECS service updated",
		UpdatedAt:    time.Now(),
	}, nil
}

// Destroy removes the ECS service for a tenant.
func (p *Provider) Destroy(ctx context.Context, tenantID string) error {
	cfg := p.getConfig(tenantID)
	if cfg == nil {
		return nil
	}

	region := resolveRegion(cfg)
	if region == "" {
		return nil
	}

	awsCfg, err := p.loadAWSConfig(ctx, awsconfig.Options{
		Region:     region,
		AssumeRole: toAssumeRoleOptions(cfg.AssumeRole),
	})
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}

	client := ecs.NewFromConfig(awsCfg)
	serviceName := resolveServiceName(cfg, tenantID)

	service, err := p.describeService(ctx, client, cfg.ClusterARN, serviceName)
	if err != nil {
		return err
	}
	if service == nil {
		p.deleteConfig(tenantID)
		return nil
	}

	_, err = client.DeleteService(ctx, &ecs.DeleteServiceInput{
		Cluster: aws.String(cfg.ClusterARN),
		Service: aws.String(serviceName),
		Force:   aws.Bool(true),
	})
	if err != nil {
		return err
	}

	p.deleteConfig(tenantID)
	return nil
}

// GetStatus queries the ECS service status for a tenant.
func (p *Provider) GetStatus(ctx context.Context, tenantID string) (*compute.ComputeStatus, error) {
	cfg := p.getConfig(tenantID)
	if cfg == nil {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	region := resolveRegion(cfg)
	if region == "" {
		return nil, fmt.Errorf("region is required for ECS provider")
	}

	awsCfg, err := p.loadAWSConfig(ctx, awsconfig.Options{
		Region:     region,
		AssumeRole: toAssumeRoleOptions(cfg.AssumeRole),
	})
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := ecs.NewFromConfig(awsCfg)
	serviceName := resolveServiceName(cfg, tenantID)

	service, err := p.describeService(ctx, client, cfg.ClusterARN, serviceName)
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	state := mapServiceState(service)
	health := compute.HealthStatusUnknown
	if state == compute.ComputeStateRunning {
		health = compute.HealthStatusHealthy
	}
	if state == compute.ComputeStateFailed {
		health = compute.HealthStatusUnhealthy
	}

	return &compute.ComputeStatus{
		TenantID:     tenantID,
		ProviderType: p.Name(),
		State:        state,
		Health:       health,
		LastUpdated:  time.Now(),
		Metadata: map[string]string{
			"service": serviceName,
			"cluster": cfg.ClusterARN,
		},
	}, nil
}

// Validate checks the compute spec for ECS requirements.
func (p *Provider) Validate(ctx context.Context, spec *compute.TenantComputeSpec) error {
	if spec == nil {
		return errors.New("spec is nil")
	}
	if spec.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if spec.ProviderType != "" && spec.ProviderType != p.Name() {
		return fmt.Errorf("provider_type must be %s", p.Name())
	}
	_, err := parseComputeConfig(spec.ProviderConfig, p.defaultConfig)
	return err
}

// ValidateConfig validates provider-specific configuration.
func (p *Provider) ValidateConfig(config json.RawMessage) error {
	_, err := parseComputeConfig(config, p.defaultConfig)
	return err
}

// ConfigSchema returns the JSON Schema for ECS compute_config.
func (p *Provider) ConfigSchema() json.RawMessage {
	return ecsConfigSchema
}

// ConfigDefaults returns no defaults for ECS compute_config.
func (p *Provider) ConfigDefaults() json.RawMessage {
	return p.defaultConfigRaw
}

func (p *Provider) describeService(ctx context.Context, client *ecs.Client, clusterARN string, serviceName string) (*ecstypes.Service, error) {
	resp, err := client.DescribeServices(ctx, &ecs.DescribeServicesInput{
		Cluster:  aws.String(clusterARN),
		Services: []string{serviceName},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Services) > 0 {
		return &resp.Services[0], nil
	}
	for _, failure := range resp.Failures {
		if failure.Reason != nil && *failure.Reason == "MISSING" {
			return nil, nil
		}
	}
	return nil, nil
}

func (p *Provider) createService(ctx context.Context, client *ecs.Client, cfg *ComputeConfig, spec *compute.TenantComputeSpec, serviceName string) error {
	desired := int32(1)
	if cfg.DesiredCount != nil {
		desired = *cfg.DesiredCount
	}

	input := &ecs.CreateServiceInput{
		Cluster:        aws.String(cfg.ClusterARN),
		ServiceName:    aws.String(serviceName),
		TaskDefinition: aws.String(cfg.TaskDefinition),
		DesiredCount:   aws.Int32(desired),
		Tags:           toECSTags(compute.MergeLabels(cfg.Tags, spec.Labels, compute.DefaultMetadata(spec))),
	}

	if cfg.LaunchType != "" {
		input.LaunchType = ecstypes.LaunchType(cfg.LaunchType)
	}

	if networkConfig := buildNetworkConfig(cfg); networkConfig != nil {
		input.NetworkConfiguration = networkConfig
	}

	_, err := client.CreateService(ctx, input)
	return err
}

func (p *Provider) updateService(ctx context.Context, client *ecs.Client, cfg *ComputeConfig, serviceName string, service *ecstypes.Service) (compute.UpdateStatus, error) {
	desired := int32(1)
	if cfg.DesiredCount != nil {
		desired = *cfg.DesiredCount
	}

	needsUpdate := false
	if service.TaskDefinition == nil || *service.TaskDefinition != cfg.TaskDefinition {
		needsUpdate = true
	}
	if service.DesiredCount != desired {
		needsUpdate = true
	}

	if !needsUpdate {
		return compute.UpdateStatusNoChanges, nil
	}

	input := &ecs.UpdateServiceInput{
		Cluster:        aws.String(cfg.ClusterARN),
		Service:        aws.String(serviceName),
		TaskDefinition: aws.String(cfg.TaskDefinition),
		DesiredCount:   aws.Int32(desired),
	}

	_, err := client.UpdateService(ctx, input)
	if err != nil {
		return compute.UpdateStatusFailed, err
	}

	return compute.UpdateStatusSuccess, nil
}

func buildNetworkConfig(cfg *ComputeConfig) *ecstypes.NetworkConfiguration {
	if cfg == nil {
		return nil
	}
	if len(cfg.Subnets) == 0 && len(cfg.SecurityGroups) == 0 && cfg.AssignPublicIP == nil {
		return nil
	}

	assignPublicIP := ecstypes.AssignPublicIpDisabled
	if cfg.AssignPublicIP != nil && *cfg.AssignPublicIP {
		assignPublicIP = ecstypes.AssignPublicIpEnabled
	}

	return &ecstypes.NetworkConfiguration{
		AwsvpcConfiguration: &ecstypes.AwsVpcConfiguration{
			Subnets:        cfg.Subnets,
			SecurityGroups: cfg.SecurityGroups,
			AssignPublicIp: assignPublicIP,
		},
	}
}

func toECSTags(tags map[string]string) []ecstypes.Tag {
	if len(tags) == 0 {
		return nil
	}

	result := make([]ecstypes.Tag, 0, len(tags))
	for key, value := range tags {
		if key == "" {
			continue
		}
		result = append(result, ecstypes.Tag{Key: aws.String(key), Value: aws.String(value)})
	}
	return result
}

func toAssumeRoleOptions(cfg *AssumeRoleConfig) *awsconfig.AssumeRoleOptions {
	if cfg == nil {
		return nil
	}
	return &awsconfig.AssumeRoleOptions{
		RoleARN:     cfg.RoleARN,
		ExternalID:  cfg.ExternalID,
		SessionName: cfg.SessionName,
	}
}

func copyConfigMap(input map[string]interface{}) map[string]interface{} {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]interface{}, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}

func marshalConfigMap(input map[string]interface{}) json.RawMessage {
	if len(input) == 0 {
		return nil
	}
	raw, err := json.Marshal(input)
	if err != nil {
		return nil
	}
	return raw
}

func mapServiceState(service *ecstypes.Service) compute.ComputeState {
	if service == nil {
		return compute.ComputeStateUnknown
	}

	if service.Deployments != nil {
		for _, deployment := range service.Deployments {
			if deployment.RolloutState == ecstypes.DeploymentRolloutStateFailed {
				return compute.ComputeStateFailed
			}
		}
	}

	status := ""
	if service.Status != nil {
		status = *service.Status
	}

	switch status {
	case "ACTIVE":
		running := service.RunningCount
		desired := service.DesiredCount
		if desired == 0 {
			return compute.ComputeStateStopped
		}
		if running == desired {
			return compute.ComputeStateRunning
		}
		return compute.ComputeStateStarting
	case "DRAINING":
		return compute.ComputeStateStopping
	case "INACTIVE":
		return compute.ComputeStateStopped
	default:
		return compute.ComputeStateUnknown
	}
}

func (p *Provider) storeConfig(tenantID string, cfg *ComputeConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tenantConfigs[tenantID] = cfg
}

func (p *Provider) getConfig(tenantID string) *ComputeConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.tenantConfigs[tenantID]
}

func (p *Provider) deleteConfig(tenantID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.tenantConfigs, tenantID)
}
