package stepfunctions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/cloud/awsconfig"
	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/workflow"
)

type Provider struct {
	region          string
	stateMachineARN string
	acct            string
	sfnClient       SFNClient
	stsClient       STSClient
	logger          *zap.Logger
}

// Config holds configuration for the Step Functions provider
type Config struct {
	Region           string
	StateMachineARN  string
	CallerAssumeRole *awsconfig.AssumeRoleOptions
}

// SFNClient is the subset of the Step Functions API used by the provider.
// It is an interface so provider tests can run without AWS credentials.
type SFNClient interface {
	StartExecution(context.Context, *sfn.StartExecutionInput, ...func(*sfn.Options)) (*sfn.StartExecutionOutput, error)
	DescribeExecution(context.Context, *sfn.DescribeExecutionInput, ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error)
	GetExecutionHistory(context.Context, *sfn.GetExecutionHistoryInput, ...func(*sfn.Options)) (*sfn.GetExecutionHistoryOutput, error)
	ListExecutions(context.Context, *sfn.ListExecutionsInput, ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error)
	StopExecution(context.Context, *sfn.StopExecutionInput, ...func(*sfn.Options)) (*sfn.StopExecutionOutput, error)
}

// STSClient is the subset of the STS API available to provider operations.
type STSClient interface {
	GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// New creates a Step Functions provider using the default AWS credential chain
// and optional caller-role assumption.
func New(ctx context.Context, cfg Config, logger *zap.Logger) (*Provider, error) {
	awsCfg, err := awsconfig.Load(ctx, awsconfig.Options{
		Region:     cfg.Region,
		AssumeRole: cfg.CallerAssumeRole,
	})
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}
	return NewWithClients(cfg, logger, sfn.NewFromConfig(awsCfg), sts.NewFromConfig(awsCfg))
}

// NewWithClients creates a provider with injected AWS clients for tests and
// alternate credential bootstrapping.
func NewWithClients(cfg Config, logger *zap.Logger, sfnClient SFNClient, stsClient STSClient) (*Provider, error) {
	if cfg.Region == "" {
		return nil, fmt.Errorf("region is required")
	}
	if cfg.StateMachineARN == "" {
		return nil, fmt.Errorf("state machine ARN is required")
	}
	if sfnClient == nil {
		return nil, fmt.Errorf("Step Functions client is required")
	}
	if stsClient == nil {
		return nil, fmt.Errorf("STS client is required")
	}
	return &Provider{
		region:          cfg.Region,
		stateMachineARN: cfg.StateMachineARN,
		sfnClient:       sfnClient,
		stsClient:       stsClient,
		logger:          logger.With(zap.String("provider", "step-functions")),
	}, nil
}

func (p *Provider) Name() string { return "step-functions" }

// Invoke starts a workflow execution using a simplified request payload
func (p *Provider) Invoke(ctx context.Context, workflowID string, request *workflow.ProvisionRequest) (*workflow.ExecutionResult, error) {
	if request == nil {
		return nil, fmt.Errorf("provision request is required")
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	executionName, err := executionName(request)
	if err != nil {
		return nil, err
	}
	input := &workflow.ExecutionInput{
		ExecutionName: executionName,
		Input:         payload,
		Tags: map[string]string{
			"tenant_id": request.TenantID,
		},
		TriggerSource: "reconciler",
	}

	return p.StartExecution(ctx, workflowID, input)
}

// GetWorkflowStatus returns a simplified workflow status for an execution
func (p *Provider) GetWorkflowStatus(ctx context.Context, executionID string) (*workflow.WorkflowStatus, error) {
	status, err := p.GetExecutionStatus(ctx, executionID)
	if err != nil {
		return nil, err
	}

	if status == nil {
		return nil, fmt.Errorf("execution status is nil")
	}

	return &workflow.WorkflowStatus{
		ExecutionID: status.ExecutionID,
		State:       status.State,
		Output:      status.Output,
		Error:       status.Error,
	}, nil
}

func (p *Provider) CreateWorkflow(ctx context.Context, spec *workflow.WorkflowSpec) (*workflow.CreateWorkflowResult, error) {
	if err := p.Validate(ctx, spec); err != nil {
		return nil, err
	}
	return &workflow.CreateWorkflowResult{
		WorkflowID:   spec.WorkflowID,
		ProviderType: "step-functions",
		ResourceIDs:  map[string]string{"arn": p.stateMachineARN},
		CreatedAt:    time.Now(),
		Message:      "using pre-provisioned state machine",
	}, nil
}

func (p *Provider) DeleteWorkflow(ctx context.Context, workflowARN string) error {
	return fmt.Errorf("delete workflow %q: configured Step Functions state machines are deployment-managed", workflowARN)
}

func (p *Provider) StartExecution(ctx context.Context, workflowID string, input *workflow.ExecutionInput) (*workflow.ExecutionResult, error) {
	if input == nil || input.ExecutionName == "" {
		return nil, fmt.Errorf("execution name is required")
	}
	if len(input.Input) == 0 || !json.Valid(input.Input) {
		return nil, fmt.Errorf("execution input must be valid JSON")
	}

	result, err := p.sfnClient.StartExecution(ctx, &sfn.StartExecutionInput{
		StateMachineArn: &p.stateMachineARN,
		Name:            &input.ExecutionName,
		Input:           stringPointer(string(input.Input)),
	})
	if err != nil {
		if isExecutionAlreadyExists(err) {
			return p.existingExecution(ctx, workflowID, input.ExecutionName, string(input.Input))
		}
		return nil, wrapAWSError(err, "start Step Functions execution")
	}
	if result.ExecutionArn == nil || *result.ExecutionArn == "" {
		return nil, fmt.Errorf("start Step Functions execution returned no execution ARN")
	}

	startedAt := time.Now()
	if result.StartDate != nil {
		startedAt = *result.StartDate
	}
	return &workflow.ExecutionResult{
		ExecutionID:  *result.ExecutionArn,
		WorkflowID:   workflowID,
		ProviderType: "step-functions",
		State:        workflow.StateRunning,
		StartedAt:    startedAt,
		Message:      "started",
	}, nil
}

func stringPointer(value string) *string { return &value }

func (p *Provider) existingExecution(ctx context.Context, workflowID, name, expectedInput string) (*workflow.ExecutionResult, error) {
	var nextToken *string
	for {
		result, err := p.sfnClient.ListExecutions(ctx, &sfn.ListExecutionsInput{
			StateMachineArn: &p.stateMachineARN,
			NextToken:       nextToken,
		})
		if err != nil {
			return nil, wrapAWSError(err, "list Step Functions executions")
		}
		for _, execution := range result.Executions {
			if execution.Name == nil || *execution.Name != name || execution.ExecutionArn == nil {
				continue
			}
			startedAt := time.Now()
			if execution.StartDate != nil {
				startedAt = *execution.StartDate
			}
			details, err := p.sfnClient.DescribeExecution(ctx, &sfn.DescribeExecutionInput{ExecutionArn: execution.ExecutionArn})
			if err != nil {
				return nil, wrapAWSError(err, "describe existing Step Functions execution")
			}
			if details.Input == nil || *details.Input != expectedInput {
				return nil, fmt.Errorf("execution %q already exists with different input", name)
			}
			return &workflow.ExecutionResult{
				ExecutionID:  *execution.ExecutionArn,
				WorkflowID:   workflowID,
				ProviderType: "step-functions",
				State:        mapExecutionState(execution.Status),
				StartedAt:    startedAt,
				Message:      "already started",
			}, nil
		}
		if result.NextToken == nil || *result.NextToken == "" {
			break
		}
		nextToken = result.NextToken
	}

	return nil, fmt.Errorf("%w: execution %q already exists but could not be resolved", workflow.ErrExecutionNotFound, name)
}

func (p *Provider) StopExecution(ctx context.Context, executionID string, reason string) error {
	if executionID == "" {
		return fmt.Errorf("execution ARN is required")
	}
	_, err := p.sfnClient.StopExecution(ctx, &sfn.StopExecutionInput{
		ExecutionArn: stringPointer(executionID),
		Error:        stringPointer("LandlordExecutionStopped"),
		Cause:        stringPointer(reason),
	})
	if err == nil || isExecutionNotFound(err) {
		return nil
	}
	return wrapAWSError(err, "stop Step Functions execution")
}

func (p *Provider) GetExecutionStatus(ctx context.Context, executionARN string) (*workflow.ExecutionStatus, error) {
	if executionARN == "" {
		return nil, fmt.Errorf("execution ARN is required")
	}

	details, err := p.sfnClient.DescribeExecution(ctx, &sfn.DescribeExecutionInput{ExecutionArn: stringPointer(executionARN)})
	if err != nil {
		return nil, wrapAWSError(err, "describe Step Functions execution")
	}
	history, err := p.executionHistory(ctx, executionARN)
	if err != nil {
		return nil, err
	}

	startTime := time.Time{}
	if details.StartDate != nil {
		startTime = *details.StartDate
	}
	status := &workflow.ExecutionStatus{
		ExecutionID:  executionARN,
		ProviderType: "step-functions",
		State:        mapExecutionState(details.Status),
		StartTime:    startTime,
		StopTime:     details.StopDate,
		Input:        rawJSON(details.Input),
		Output:       rawJSON(details.Output),
		History:      history,
	}
	if details.StateMachineArn != nil {
		status.WorkflowID = *details.StateMachineArn
	}
	if details.Error != nil || details.Cause != nil {
		status.Error = &workflow.ExecutionError{
			Code:    stringValue(details.Error),
			Message: stringValue(details.Cause),
			Cause:   stringValue(details.Cause),
		}
		if details.StopDate != nil {
			status.Error.FailedAt = details.StopDate.UTC().Format(time.RFC3339)
		}
	}
	return status, nil
}

func (p *Provider) executionHistory(ctx context.Context, executionARN string) ([]workflow.ExecutionEvent, error) {
	var events []workflow.ExecutionEvent
	var nextToken *string
	for {
		result, err := p.sfnClient.GetExecutionHistory(ctx, &sfn.GetExecutionHistoryInput{
			ExecutionArn: stringPointer(executionARN),
			NextToken:    nextToken,
			ReverseOrder: false,
		})
		if err != nil {
			return nil, wrapAWSError(err, "get Step Functions execution history")
		}
		for _, event := range result.Events {
			details, err := json.Marshal(event)
			if err != nil {
				return nil, fmt.Errorf("marshal Step Functions history event: %w", err)
			}
			timestamp := time.Time{}
			if event.Timestamp != nil {
				timestamp = *event.Timestamp
			}
			events = append(events, workflow.ExecutionEvent{
				Timestamp: timestamp,
				Type:      string(event.Type),
				Details:   details,
			})
		}
		if result.NextToken == nil || *result.NextToken == "" {
			return events, nil
		}
		nextToken = result.NextToken
	}
}

func rawJSON(value *string) json.RawMessage {
	if value == nil || !json.Valid([]byte(*value)) {
		return nil
	}
	return json.RawMessage(*value)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (p *Provider) Validate(ctx context.Context, spec *workflow.WorkflowSpec) error {
	// Basic validation: ensure workflow ID and definition exist
	if spec == nil || spec.WorkflowID == "" || len(spec.Definition) == 0 {
		return workflow.ErrInvalidSpec
	}
	return nil
}

// PostComputeCallback sends a compute execution callback to a running Step Functions execution
func (p *Provider) PostComputeCallback(ctx context.Context, executionID string, payload *compute.CallbackPayload, opts *compute.CallbackOptions) error {
	_ = executionID
	_ = opts
	// Marshal the callback payload to JSON for logging/debug
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal callback payload: %w", err)
	}
	p.logger.Debug("posting compute callback",
		zap.String("execution_id", executionID),
		zap.String("tenant_id", payload.TenantID),
		zap.ByteString("payload", payloadJSON),
	)
	return nil
}
