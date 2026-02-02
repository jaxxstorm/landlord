package compute

import (
	"encoding/json"
	"time"
)

// TenantComputeSpec defines the desired state of tenant compute resources
type TenantComputeSpec struct {
	// TenantID uniquely identifies this tenant
	TenantID string `json:"tenant_id"`

	// ProviderType specifies which compute provider to use
	// Must match a registered provider name
	ProviderType string `json:"provider_type"`

	// Containers defines the container workloads for this tenant
	Containers []ContainerSpec `json:"containers"`

	// Resources specifies compute resource requirements
	Resources ResourceRequirements `json:"resources"`

	// Networking defines network configuration
	Networking NetworkConfig `json:"networking"`

	// Secrets references secrets needed by containers
	Secrets []SecretReference `json:"secrets,omitempty"`

	// Labels are key-value pairs for organizing/filtering
	Labels map[string]string `json:"labels,omitempty"`

	// ProviderConfig contains provider-specific configuration
	// Validated by the specific provider
	ProviderConfig json.RawMessage `json:"provider_config,omitempty"`
}

// ContainerSpec defines a single container in the tenant deployment
type ContainerSpec struct {
	// Name is the container identifier within the tenant
	Name string `json:"name"`

	// Image is the container image reference (e.g., "nginx:1.21")
	Image string `json:"image"`

	// Command overrides the container's default command
	Command []string `json:"command,omitempty"`

	// Args provides arguments to the command
	Args []string `json:"args,omitempty"`

	// Env defines environment variables
	Env map[string]string `json:"env,omitempty"`

	// Ports defines port mappings
	Ports []PortMapping `json:"ports,omitempty"`

	// HealthCheck configures container health checking
	HealthCheck *HealthCheckConfig `json:"health_check,omitempty"`

	// Resources specifies container-specific resource limits
	Resources *ResourceRequirements `json:"resources,omitempty"`
}

// PortMapping defines how a container port is exposed
type PortMapping struct {
	// ContainerPort is the port inside the container
	ContainerPort int `json:"container_port"`

	// HostPort is the port on the host (optional, provider-specific)
	HostPort int `json:"host_port,omitempty"`

	// Protocol is tcp or udp
	Protocol string `json:"protocol"` // "tcp" or "udp"

	// Name is an optional identifier for this port
	Name string `json:"name,omitempty"`
}

// HealthCheckConfig defines health checking parameters
type HealthCheckConfig struct {
	// Type is the health check method
	Type string `json:"type"` // "http", "tcp", "exec"

	// HTTPPath for HTTP health checks
	HTTPPath string `json:"http_path,omitempty"`

	// Port for HTTP/TCP health checks
	Port int `json:"port,omitempty"`

	// Command for exec health checks
	Command []string `json:"command,omitempty"`

	// IntervalSeconds between health checks
	IntervalSeconds int `json:"interval_seconds"`

	// TimeoutSeconds for each health check
	TimeoutSeconds int `json:"timeout_seconds"`

	// HealthyThreshold consecutive successes needed
	HealthyThreshold int `json:"healthy_threshold"`

	// UnhealthyThreshold consecutive failures needed
	UnhealthyThreshold int `json:"unhealthy_threshold"`
}

// ResourceRequirements defines compute resource limits
type ResourceRequirements struct {
	// CPU in millicores (1000 = 1 CPU)
	CPU int `json:"cpu"`

	// Memory in megabytes
	Memory int `json:"memory"`

	// Storage in megabytes (ephemeral)
	Storage int `json:"storage,omitempty"`
}

// NetworkConfig defines networking parameters
type NetworkConfig struct {
	// PublicIP whether to assign a public IP
	PublicIP bool `json:"public_ip"`

	// SecurityGroups or firewall rules
	SecurityGroups []string `json:"security_groups,omitempty"`

	// Subnets to place resources in
	Subnets []string `json:"subnets,omitempty"`

	// DNS settings
	DNS *DNSConfig `json:"dns,omitempty"`
}

// DNSConfig defines DNS settings
type DNSConfig struct {
	// Hostname for the tenant
	Hostname string `json:"hostname,omitempty"`

	// Domain for the tenant
	Domain string `json:"domain,omitempty"`
}

// SecretReference points to a secret needed by the tenant
type SecretReference struct {
	// Name of the secret
	Name string `json:"name"`

	// Source where the secret is stored (provider-specific)
	Source string `json:"source"` // e.g., "aws-secrets-manager", "kubernetes-secret"

	// Key within the secret store
	Key string `json:"key"`

	// EnvVar to inject the secret as
	EnvVar string `json:"env_var,omitempty"`
}

// ProvisionResult contains the outcome of a provision operation
type ProvisionResult struct {
	// TenantID that was provisioned
	TenantID string `json:"tenant_id"`

	// ProviderType used for provisioning
	ProviderType string `json:"provider_type"`

	// Status of the provisioning operation
	Status ProvisionStatus `json:"status"`

	// ResourceIDs maps resource types to provider-specific IDs
	// Examples: {"task": "arn:...", "service": "arn:..."}
	ResourceIDs map[string]string `json:"resource_ids"`

	// Endpoints where the tenant is accessible
	Endpoints []Endpoint `json:"endpoints,omitempty"`

	// Message provides additional details
	Message string `json:"message,omitempty"`

	// ProvisionedAt timestamp
	ProvisionedAt time.Time `json:"provisioned_at"`
}

// ProvisionStatus indicates provisioning outcome
type ProvisionStatus string

const (
	ProvisionStatusSuccess    ProvisionStatus = "success"
	ProvisionStatusInProgress ProvisionStatus = "in_progress"
	ProvisionStatusFailed     ProvisionStatus = "failed"
)

// UpdateResult contains the outcome of an update operation
type UpdateResult struct {
	// TenantID that was updated
	TenantID string `json:"tenant_id"`

	// ProviderType used
	ProviderType string `json:"provider_type"`

	// Status of the update
	Status UpdateStatus `json:"status"`

	// Changes describes what was modified
	Changes []string `json:"changes"`

	// Message provides additional details
	Message string `json:"message,omitempty"`

	// UpdatedAt timestamp
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateStatus indicates update outcome
type UpdateStatus string

const (
	UpdateStatusSuccess    UpdateStatus = "success"
	UpdateStatusNoChanges  UpdateStatus = "no_changes"
	UpdateStatusInProgress UpdateStatus = "in_progress"
	UpdateStatusFailed     UpdateStatus = "failed"
)

// Endpoint describes how to reach a tenant
type Endpoint struct {
	// Type of endpoint
	Type string `json:"type"` // "http", "https", "tcp", "grpc"

	// Address (hostname or IP)
	Address string `json:"address"`

	// Port number
	Port int `json:"port"`

	// URL if applicable
	URL string `json:"url,omitempty"`
}

// ComputeStatus represents current state of tenant compute
type ComputeStatus struct {
	// TenantID being queried
	TenantID string `json:"tenant_id"`

	// ProviderType managing this tenant
	ProviderType string `json:"provider_type"`

	// State of the overall deployment
	State ComputeState `json:"state"`

	// Containers shows status of each container
	Containers []ContainerStatus `json:"containers"`

	// Health overall health status
	Health HealthStatus `json:"health"`

	// LastUpdated when this status was checked
	LastUpdated time.Time `json:"last_updated"`

	// Metadata provider-specific status information
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ComputeState represents deployment state
type ComputeState string

const (
	ComputeStateRunning  ComputeState = "running"
	ComputeStateStopped  ComputeState = "stopped"
	ComputeStateStarting ComputeState = "starting"
	ComputeStateStopping ComputeState = "stopping"
	ComputeStateFailed   ComputeState = "failed"
	ComputeStateUnknown  ComputeState = "unknown"
)

// ContainerStatus shows state of a single container
type ContainerStatus struct {
	// Name of the container
	Name string `json:"name"`

	// State of this container
	State string `json:"state"` // "running", "stopped", "failed"

	// Ready whether container is ready to serve
	Ready bool `json:"ready"`

	// RestartCount how many times restarted
	RestartCount int `json:"restart_count"`

	// Message additional details
	Message string `json:"message,omitempty"`
}

// HealthStatus overall health assessment
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ComputeExecutionStatus represents the status of a compute provisioning operation
type ComputeExecutionStatus string

const (
	ExecutionStatusPending   ComputeExecutionStatus = "pending"
	ExecutionStatusRunning   ComputeExecutionStatus = "running"
	ExecutionStatusSucceeded ComputeExecutionStatus = "succeeded"
	ExecutionStatusFailed    ComputeExecutionStatus = "failed"
)

// ComputeOperationType represents the type of compute operation
type ComputeOperationType string

const (
	OperationTypeProvision ComputeOperationType = "provision"
	OperationTypeUpdate    ComputeOperationType = "update"
	OperationTypeDelete    ComputeOperationType = "delete"
)

// ComputeExecution tracks a single compute provisioning operation
type ComputeExecution struct {
	// ID is the unique database identifier
	ID int64 `db:"id"`

	// ExecutionID is a deterministic unique identifier for this operation
	ExecutionID string `db:"execution_id" json:"execution_id"`

	// TenantID links this execution to the tenant
	TenantID string `db:"tenant_id" json:"tenant_id"`

	// WorkflowExecutionID links back to the triggering workflow
	WorkflowExecutionID string `db:"workflow_execution_id" json:"workflow_execution_id"`

	// OperationType is the kind of operation (provision, update, delete)
	OperationType ComputeOperationType `db:"operation_type" json:"operation_type"`

	// Status tracks the current state (pending, running, succeeded, failed)
	Status ComputeExecutionStatus `db:"status" json:"status"`

	// ResourceIDs contains provider-specific resource identifiers
	ResourceIDs json.RawMessage `db:"resource_ids" json:"resource_ids,omitempty"`

	// ErrorCode is populated if status is "failed"
	ErrorCode *string `db:"error_code" json:"error_code,omitempty"`

	// ErrorMessage provides details about the failure
	ErrorMessage *string `db:"error_message" json:"error_message,omitempty"`

	// CreatedAt is when the execution was initiated
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	// UpdatedAt is when the execution was last updated
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ComputeExecutionHistory represents a state transition in a compute execution
type ComputeExecutionHistory struct {
	// ID is the unique database identifier
	ID int64 `db:"id"`

	// ComputeExecutionID links to the parent compute execution
	ComputeExecutionID string `db:"compute_execution_id" json:"compute_execution_id"`

	// Status is the state at this history entry
	Status ComputeExecutionStatus `db:"status" json:"status"`

	// Details contains arbitrary details about this state transition
	Details json.RawMessage `db:"details" json:"details,omitempty"`

	// Timestamp is when this state transition occurred
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
}

// ComputeError represents a standardized error from compute operations
type ComputeError struct {
	// Code is a standard error code for classification
	Code string `json:"code"`

	// Message is a human-readable error description
	Message string `json:"message"`

	// IsRetriable indicates whether the operation can be retried
	IsRetriable bool `json:"is_retriable"`

	// ProviderError contains the raw error from the provider for debugging
	ProviderError string `json:"provider_error,omitempty"`
}

// CallbackPayload represents a callback from compute operation completion
type CallbackPayload struct {
	// ExecutionID is the compute execution ID that completed
	ExecutionID string `json:"execution_id"`

	// TenantID is the associated tenant
	TenantID string `json:"tenant_id"`

	// Status is the final status of the compute operation
	Status ComputeExecutionStatus `json:"status"`

	// ResourceIDs contains created resources (for successful operations)
	ResourceIDs map[string]interface{} `json:"resource_ids,omitempty"`

	// ErrorCode is populated for failed operations
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is populated for failed operations
	ErrorMessage string `json:"error_message,omitempty"`

	// IsRetriable indicates if the operation can be retried
	IsRetriable bool `json:"is_retriable"`
}

// CallbackOptions controls callback delivery behavior
type CallbackOptions struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryDelay is the initial delay between retries
	RetryDelay time.Duration

	// BackoffType defines the retry backoff strategy (exponential or linear)
	BackoffType string
}

// FailedCallback represents a callback that failed delivery
type FailedCallback struct {
	// ExecutionID is the compute execution ID
	ExecutionID string `json:"execution_id"`

	// Payload is the callback payload that failed to deliver
	Payload *CallbackPayload `json:"payload"`

	// Error is the last error encountered during delivery
	Error string `json:"error"`

	// Attempts is the number of delivery attempts made
	Attempts int `json:"attempts"`

	// FailedAt is when the callback first failed
	FailedAt time.Time `json:"failed_at"`

	// LastAttemptAt is when the last delivery attempt was made
	LastAttemptAt time.Time `json:"last_attempt_at"`
}
