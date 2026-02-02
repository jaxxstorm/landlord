# Spec: Compute Types and Data Structures

## Overview

Defines the data structures used across the compute provisioning framework.

## Tenant Compute Specification

```go
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
```

## Result Types

```go
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
```

## Status Types

```go
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
    ComputeStateRunning     ComputeState = "running"
    ComputeStateStopped     ComputeState = "stopped"
    ComputeStateStarting    ComputeState = "starting"
    ComputeStateStopping    ComputeState = "stopping"
    ComputeStateFailed      ComputeState = "failed"
    ComputeStateUnknown     ComputeState = "unknown"
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
```

## Validation Rules

### TenantComputeSpec
- `TenantID` **MUST** be non-empty and match pattern `^[a-z0-9-]+$`
- `ProviderType` **MUST** be non-empty and match a registered provider
- `Containers` **MUST** have at least one container
- `Containers[].Name` **MUST** be unique within the spec

### ContainerSpec
- `Name` **MUST** match pattern `^[a-z0-9-]+$`
- `Image` **MUST** be a valid image reference
- `Ports[].ContainerPort` **MUST** be 1-65535
- `Ports[].Protocol` **MUST** be "tcp" or "udp"

### ResourceRequirements
- `CPU` **MUST** be >= 128 (0.128 CPU)
- `Memory` **MUST** be >= 128 (128 MB)

### HealthCheckConfig
- `Type` **MUST** be "http", "tcp", or "exec"
- `IntervalSeconds` **MUST** be >= 5
- `TimeoutSeconds` **MUST** be >= 1
- `TimeoutSeconds` **MUST** be < `IntervalSeconds`
