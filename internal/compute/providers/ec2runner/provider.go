package ec2runner

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/cloud/awsconfig"
	"github.com/jaxxstorm/landlord/internal/compute"
)

// Config contains operator-owned EC2 launch defaults.
type Config struct {
	Region           string
	InstanceType     string
	SubnetID         string
	SecurityGroupIDs []string
	InstanceProfile  string
	KeyName          string
	AssumeRole       *awsconfig.AssumeRoleOptions
}

// Provider manages one EC2 runner instance per tenant.
type Provider struct {
	config Config
	client EC2Client
	logger *zap.Logger
}

// New creates an EC2 runner provider using the standard AWS credential chain.
func New(ctx context.Context, cfg Config, logger *zap.Logger) (*Provider, error) {
	awsCfg, err := awsconfig.Load(ctx, awsconfig.Options{Region: cfg.Region, AssumeRole: cfg.AssumeRole})
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}
	return NewWithClient(cfg, logger, ec2.NewFromConfig(awsCfg))
}

// NewWithClient creates an EC2 runner provider with an injected client for tests.
func NewWithClient(cfg Config, logger *zap.Logger, client EC2Client) (*Provider, error) {
	if err := validateOperatorConfig(cfg); err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("EC2 client is required")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Provider{
		config: cfg,
		client: client,
		logger: logger.With(zap.String("component", "ec2-runner-provider")),
	}, nil
}

func validateOperatorConfig(cfg Config) error {
	if strings.TrimSpace(cfg.Region) == "" {
		return fmt.Errorf("region is required")
	}
	if strings.TrimSpace(cfg.InstanceType) == "" {
		return fmt.Errorf("instance type is required")
	}
	if strings.TrimSpace(cfg.SubnetID) == "" {
		return fmt.Errorf("subnet ID is required")
	}
	if len(cfg.SecurityGroupIDs) == 0 {
		return fmt.Errorf("security group IDs are required")
	}
	for _, id := range cfg.SecurityGroupIDs {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("security group IDs must not contain empty values")
		}
	}
	return nil
}

// Name returns the compute provider identifier.
func (p *Provider) Name() string { return "ec2-runner" }

// ValidateConfig validates tenant-facing AMI and user-data configuration.
func (p *Provider) ValidateConfig(raw json.RawMessage) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return fmt.Errorf("invalid EC2 runner config: %w", err)
	}
	for field := range fields {
		if field != "ami" && field != "user_data" {
			return fmt.Errorf("unsupported EC2 runner config field: %s", field)
		}
	}
	_, err := parseLaunchConfig(raw)
	return err
}

// ConfigSchema returns the tenant-facing EC2 runner configuration schema.
func (p *Provider) ConfigSchema() json.RawMessage { return launchConfigSchema }

// ConfigDefaults omits operator-owned launch defaults from tenant discovery.
func (p *Provider) ConfigDefaults() json.RawMessage { return nil }

// Validate validates an EC2 runner compute specification without mutation.
func (p *Provider) Validate(_ context.Context, spec *compute.TenantComputeSpec) error {
	if spec == nil || strings.TrimSpace(spec.TenantID) == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if spec.ProviderType != "" && spec.ProviderType != p.Name() {
		return fmt.Errorf("provider type must be %s", p.Name())
	}
	return p.ValidateConfig(spec.ProviderConfig)
}

// Provision creates a runner unless one is already managed for the tenant.
func (p *Provider) Provision(ctx context.Context, spec *compute.TenantComputeSpec) (*compute.ProvisionResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := p.Validate(ctx, spec); err != nil {
		return nil, fmt.Errorf("%w: %v", compute.ErrInvalidSpec, err)
	}
	config, err := parseLaunchConfig(spec.ProviderConfig)
	if err != nil {
		return nil, err
	}
	compute.ApplyDefaultMetadata(spec)
	existing, err := p.findActiveInstances(ctx, spec.TenantID)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("describe EC2 runners: %w", err)
	}
	if len(existing) > 0 {
		return provisionResult(spec.TenantID, existing[0]), nil
	}

	instance, err := p.launchInstance(ctx, spec, config)
	if err != nil {
		return nil, err
	}
	return provisionResult(spec.TenantID, instance), nil
}

func (p *Provider) launchInstance(ctx context.Context, spec *compute.TenantComputeSpec, config *LaunchConfig) (ec2types.Instance, error) {
	input := &ec2.RunInstancesInput{
		ImageId:          stringPointer(config.AMI),
		InstanceType:     ec2types.InstanceType(p.config.InstanceType),
		MinCount:         int32Pointer(1),
		MaxCount:         int32Pointer(1),
		SubnetId:         stringPointer(p.config.SubnetID),
		SecurityGroupIds: p.config.SecurityGroupIDs,
		UserData:         stringPointer(base64.StdEncoding.EncodeToString([]byte(config.UserData))),
		TagSpecifications: []ec2types.TagSpecification{{
			ResourceType: ec2types.ResourceTypeInstance,
			Tags:         managedTags(spec, config),
		}},
	}
	if p.config.InstanceProfile != "" {
		input.IamInstanceProfile = &ec2types.IamInstanceProfileSpecification{Name: stringPointer(p.config.InstanceProfile)}
	}
	if p.config.KeyName != "" {
		input.KeyName = stringPointer(p.config.KeyName)
	}
	result, err := p.client.RunInstances(ctx, input)
	if err != nil {
		if ctx.Err() != nil {
			return ec2types.Instance{}, ctx.Err()
		}
		return ec2types.Instance{}, fmt.Errorf("%w: launch EC2 runner: %v", compute.ErrProvisionFailed, err)
	}
	if len(result.Instances) == 0 {
		return ec2types.Instance{}, fmt.Errorf("%w: launch EC2 runner returned no instances", compute.ErrProvisionFailed)
	}
	return result.Instances[0], nil
}

// Update leaves an unchanged runner intact and replaces it when AMI or user
// data changes. The new instance is launched before the old one is terminated.
func (p *Provider) Update(ctx context.Context, tenantID string, spec *compute.TenantComputeSpec) (*compute.UpdateResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if spec == nil {
		return nil, fmt.Errorf("%w: spec is required", compute.ErrInvalidSpec)
	}
	copy := *spec
	copy.TenantID = tenantID
	if err := p.Validate(ctx, &copy); err != nil {
		return nil, fmt.Errorf("%w: %v", compute.ErrInvalidSpec, err)
	}
	config, err := parseLaunchConfig(copy.ProviderConfig)
	if err != nil {
		return nil, err
	}
	compute.ApplyDefaultMetadata(&copy)
	existing, err := p.findActiveInstances(ctx, tenantID)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("describe EC2 runners: %w", err)
	}
	if len(existing) == 0 {
		if _, err := p.launchInstance(ctx, &copy, config); err != nil {
			return nil, err
		}
		return updateResult(tenantID, compute.UpdateStatusSuccess, []string{"runner created"}), nil
	}
	if instanceFingerprint(existing[0]) == launchFingerprint(config) {
		return updateResult(tenantID, compute.UpdateStatusNoChanges, nil), nil
	}
	if _, err := p.launchInstance(ctx, &copy, config); err != nil {
		return nil, err
	}
	if existing[0].InstanceId != nil {
		if _, err := p.client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{InstanceIds: []string{*existing[0].InstanceId}}); err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, fmt.Errorf("terminate replaced EC2 runner: %w", err)
		}
	}
	return updateResult(tenantID, compute.UpdateStatusSuccess, []string{"runner replaced"}), nil
}

func updateResult(tenantID string, status compute.UpdateStatus, changes []string) *compute.UpdateResult {
	return &compute.UpdateResult{TenantID: tenantID, ProviderType: "ec2-runner", Status: status, Changes: changes, UpdatedAt: time.Now()}
}

func instanceFingerprint(instance ec2types.Instance) string {
	for _, tag := range instance.Tags {
		if tag.Key != nil && *tag.Key == launchFingerprintTag && tag.Value != nil {
			return *tag.Value
		}
	}
	return ""
}

// GetStatus reports the lifecycle state and addresses of a managed EC2 runner.
func (p *Provider) GetStatus(ctx context.Context, tenantID string) (*compute.ComputeStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	instances, err := p.findActiveInstances(ctx, tenantID)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("describe EC2 runners: %w", err)
	}
	if len(instances) == 0 {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}
	instance := instances[0]
	state := mapInstanceState(instance)
	health := compute.HealthStatusUnknown
	if state == compute.ComputeStateRunning {
		health = compute.HealthStatusHealthy
	}
	if state == compute.ComputeStateFailed {
		health = compute.HealthStatusUnhealthy
	}
	metadata := map[string]string{}
	if instance.InstanceId != nil {
		metadata["instance"] = *instance.InstanceId
	}
	if instance.PrivateIpAddress != nil {
		metadata["private_ip"] = *instance.PrivateIpAddress
	}
	if instance.PublicIpAddress != nil {
		metadata["public_ip"] = *instance.PublicIpAddress
	}
	return &compute.ComputeStatus{
		TenantID: tenantID, ProviderType: p.Name(), State: state, Health: health,
		LastUpdated: time.Now(), Metadata: metadata,
	}, nil
}

// Destroy terminates managed EC2 runners and is a no-op when none remain.
func (p *Provider) Destroy(ctx context.Context, tenantID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	instances, err := p.findActiveInstances(ctx, tenantID)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("describe EC2 runners: %w", err)
	}
	instanceIDs := make([]string, 0, len(instances))
	for _, instance := range instances {
		if instance.InstanceId != nil && *instance.InstanceId != "" {
			instanceIDs = append(instanceIDs, *instance.InstanceId)
		}
	}
	if len(instanceIDs) == 0 {
		return nil
	}
	if _, err := p.client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{InstanceIds: instanceIDs}); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("terminate EC2 runners: %w", err)
	}
	return nil
}

func mapInstanceState(instance ec2types.Instance) compute.ComputeState {
	if instance.State == nil {
		return compute.ComputeStateUnknown
	}
	switch instance.State.Name {
	case ec2types.InstanceStateNamePending:
		return compute.ComputeStateStarting
	case ec2types.InstanceStateNameRunning:
		return compute.ComputeStateRunning
	case ec2types.InstanceStateNameStopping:
		return compute.ComputeStateStopping
	case ec2types.InstanceStateNameStopped:
		return compute.ComputeStateStopped
	case ec2types.InstanceStateNameShuttingDown, ec2types.InstanceStateNameTerminated:
		return compute.ComputeStateFailed
	default:
		return compute.ComputeStateUnknown
	}
}

func (p *Provider) findActiveInstances(ctx context.Context, tenantID string) ([]ec2types.Instance, error) {
	filters := []ec2types.Filter{
		{Name: stringPointer("tag:" + compute.MetadataOwnerKey), Values: []string{compute.MetadataOwnerValue}},
		{Name: stringPointer("tag:" + compute.MetadataTenantIDKey), Values: []string{tenantID}},
		{Name: stringPointer("tag:" + compute.MetadataProviderKey), Values: []string{p.Name()}},
	}
	result, err := p.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{Filters: filters})
	if err != nil {
		return nil, err
	}
	instances := make([]ec2types.Instance, 0)
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if instance.State != nil && instance.State.Name == ec2types.InstanceStateNameTerminated {
				continue
			}
			instances = append(instances, instance)
		}
	}
	return instances, nil
}

func provisionResult(tenantID string, instance ec2types.Instance) *compute.ProvisionResult {
	resourceIDs := map[string]string{}
	if instance.InstanceId != nil {
		resourceIDs["instance"] = *instance.InstanceId
	}
	return &compute.ProvisionResult{
		TenantID:      tenantID,
		ProviderType:  "ec2-runner",
		Status:        compute.ProvisionStatusSuccess,
		ResourceIDs:   resourceIDs,
		Endpoints:     instanceEndpoints(instance),
		Message:       "EC2 runner provisioned",
		ProvisionedAt: time.Now(),
	}
}

func instanceEndpoints(instance ec2types.Instance) []compute.Endpoint {
	endpoints := make([]compute.Endpoint, 0, 2)
	if instance.PrivateIpAddress != nil && *instance.PrivateIpAddress != "" {
		endpoints = append(endpoints, compute.Endpoint{Type: "private-ip", Address: *instance.PrivateIpAddress})
	}
	if instance.PublicIpAddress != nil && *instance.PublicIpAddress != "" {
		endpoints = append(endpoints, compute.Endpoint{Type: "public-ip", Address: *instance.PublicIpAddress})
	}
	return endpoints
}

func stringPointer(value string) *string { return &value }

func int32Pointer(value int32) *int32 { return &value }
