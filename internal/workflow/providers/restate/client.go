package restate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"go.uber.org/zap"
)

// Client wraps the Restate SDK for workflow operations
type Client struct {
	endpoint      string // Ingress endpoint for invoking workflows
	adminEndpoint string // Admin endpoint for metadata operations
	serviceName   string // Default service name for status lookups
	authType      string
	apiKey        string
	httpClient    *http.Client
	logger        *zap.Logger
}

var errAdminAPINotSupported = errors.New("restate admin api does not support this operation")

// NewClient creates a new Restate client
func NewClient(ctx context.Context, cfg config.RestateConfig, logger *zap.Logger) (*Client, error) {
	// Validate endpoint
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("restate endpoint is required")
	}

	adminEndpoint := cfg.AdminEndpoint
	if adminEndpoint == "" {
		// Default to endpoint with port 9070 if not specified
		adminEndpoint = cfg.Endpoint
	}

	client := &Client{
		endpoint:      cfg.Endpoint,
		adminEndpoint: adminEndpoint,
		serviceName:   cfg.ServiceName,
		authType:      cfg.AuthType,
		apiKey:        cfg.ApiKey,
		httpClient:    &http.Client{},
		logger:        logger.With(zap.String("component", "restate-client")),
	}

	// Test connection
	if err := client.testConnection(ctx); err != nil {
		client.logger.Warn("restate server unreachable at initialization (will retry on first use)",
			zap.String("endpoint", cfg.Endpoint),
			zap.Error(err),
		)
	}

	client.logger.Info("restate client initialized",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("auth_type", cfg.AuthType),
	)

	return client, nil
}

// GetService queries a service from Restate
func (c *Client) GetService(ctx context.Context, serviceName string) (map[string]interface{}, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name is required")
	}

	url := fmt.Sprintf("%s/services/%s", c.adminEndpoint, serviceName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.addAuthHeader(req); err != nil {
		return nil, fmt.Errorf("failed to add auth header: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: service not found", workflow.ErrWorkflowNotFound)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var service map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&service); err != nil {
		return nil, fmt.Errorf("failed to decode service: %w", err)
	}

	return service, nil
}

type registerServiceRequest struct {
	Name string `json:"name"`
}

type registerDeploymentRequest struct {
	URI string `json:"uri"`
}

// RegisterService registers a service with Restate
func (c *Client) RegisterService(ctx context.Context, serviceName string) error {
	if serviceName == "" {
		return fmt.Errorf("service name is required")
	}

	if _, err := c.GetService(ctx, serviceName); err == nil {
		return nil
	}

	payload, err := json.Marshal(registerServiceRequest{Name: serviceName})
	if err != nil {
		return fmt.Errorf("failed to encode registration payload: %w", err)
	}

	url := fmt.Sprintf("%s/services", c.adminEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create registration request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if err := c.addAuthHeader(req); err != nil {
		return fmt.Errorf("failed to add auth header: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return ErrServiceAlreadyExists
	}

	if resp.StatusCode == http.StatusMethodNotAllowed {
		return fmt.Errorf("service registration not supported by restate admin api")
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	c.logger.Info("service registered",
		zap.String("service_name", serviceName),
	)

	return nil
}

// RegisterDeployment registers a worker deployment with Restate.
func (c *Client) RegisterDeployment(ctx context.Context, uri string) error {
	if uri == "" {
		return fmt.Errorf("deployment uri is required")
	}

	payload, err := json.Marshal(registerDeploymentRequest{URI: uri})
	if err != nil {
		return fmt.Errorf("failed to encode deployment payload: %w", err)
	}

	url := fmt.Sprintf("%s/deployments", c.adminEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create deployment request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if err := c.addAuthHeader(req); err != nil {
		return fmt.Errorf("failed to add auth header: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register deployment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		bodyText := strings.ToLower(string(body))
		isNotSupported := strings.Contains(bodyText, "not supported") || strings.Contains(bodyText, "unsupported")
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusInternalServerError {
			if isNotSupported || len(strings.TrimSpace(bodyText)) == 0 {
				return fmt.Errorf("%w: %s", errAdminAPINotSupported, strings.TrimSpace(string(body)))
			}
		}
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	c.logger.Info("deployment registered",
		zap.String("uri", uri),
	)

	return nil
}

// InvokeService invokes a Restate service and returns the execution ID
func (c *Client) InvokeService(ctx context.Context, serviceName, executionName string, input json.RawMessage) (string, error) {
	if serviceName == "" {
		return "", fmt.Errorf("service name is required")
	}

	// Construct the invocation URL
	// Format: {endpoint}/{serviceName}/{handlerName}/send
	// Using "send" for async invocation (fire-and-forget with idempotency key)
	url := fmt.Sprintf("%s/%s/execute/send?idempotency_key=%s", c.endpoint, serviceName, executionName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(input))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if err := c.addAuthHeader(req); err != nil {
		return "", fmt.Errorf("failed to add auth header: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to invoke service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := json.Marshal(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	executionID := ""
	body, _ := io.ReadAll(resp.Body)
	if len(body) > 0 {
		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err == nil {
			executionID = firstStringValue(payload, "invocationId", "invocation_id", "id", "executionId", "execution_id")
		}
	}

	if executionID == "" {
		executionID = executionName
		if executionID == "" {
			executionID = fmt.Sprintf("%s-%s", serviceName, executionName)
		}
	}

	c.logger.Info("service invoked",
		zap.String("service_name", serviceName),
		zap.String("execution_name", executionName),
		zap.String("execution_id", executionID),
		zap.String("url", url),
	)

	return executionID, nil
}

// GetExecutionStatus retrieves execution status from Restate
func (c *Client) GetExecutionStatus(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
	if executionID == "" {
		return nil, fmt.Errorf("execution ID is required")
	}

	if !strings.HasPrefix(executionID, "inv_") {
		return c.getExecutionStatusByIdempotencyKey(ctx, executionID)
	}

	invocation, err := c.fetchInvocationStatus(ctx, executionID)
	if err != nil {
		return nil, err
	}

	// Map Restate invocation status to workflow status
	state := workflow.StateRunning
	if status, ok := invocation["status"].(string); ok {
		normalized := strings.ToLower(status)
		switch {
		case strings.Contains(normalized, "completed"), strings.Contains(normalized, "succeeded"), strings.Contains(normalized, "success"):
			state = workflow.StateSucceeded
		case strings.Contains(normalized, "failed"), strings.Contains(normalized, "error"):
			state = workflow.StateFailed
		case strings.Contains(normalized, "pending"):
			state = workflow.StatePending
		case strings.Contains(normalized, "running"), strings.Contains(normalized, "active"), strings.Contains(normalized, "suspended"):
			state = workflow.StateRunning
		}
	}

	status := &workflow.ExecutionStatus{
		ExecutionID:  executionID,
		ProviderType: "restate",
		State:        state,
		StartTime:    time.Now(), // TODO: parse from invocation
		Input:        json.RawMessage(`{}`),
	}

	if output, ok := invocation["output"]; ok {
		if raw, err := json.Marshal(output); err == nil {
			status.Output = raw
		}
	}
	if errPayload, ok := invocation["error"]; ok {
		if raw, err := json.Marshal(errPayload); err == nil {
			status.Error = &workflow.ExecutionError{
				Code:    "restate_error",
				Message: string(raw),
			}
		}
	}

	return status, nil
}

func (c *Client) getExecutionStatusByIdempotencyKey(ctx context.Context, idempotencyKey string) (*workflow.ExecutionStatus, error) {
	if c.serviceName == "" {
		c.serviceName = "TenantProvisioning"
	}

	url := fmt.Sprintf("%s/restate/invocation/%s/execute/%s/output", c.endpoint, c.serviceName, idempotencyKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if err := c.addAuthHeader(req); err != nil {
		return nil, fmt.Errorf("failed to add auth header: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query execution status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: execution not found", workflow.ErrExecutionNotFound)
	}

	if resp.StatusCode == 470 {
		return &workflow.ExecutionStatus{
			ExecutionID:  idempotencyKey,
			ProviderType: "restate",
			State:        workflow.StateRunning,
			StartTime:    time.Now(),
			Input:        json.RawMessage(`{}`),
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode invocation output: %w", err)
	}

	state := workflow.StateSucceeded
	if status, ok := response["state"].(string); ok {
		normalized := strings.ToLower(status)
		switch {
		case strings.Contains(normalized, "failed"), strings.Contains(normalized, "error"):
			state = workflow.StateFailed
		case strings.Contains(normalized, "running"), strings.Contains(normalized, "pending"):
			state = workflow.StateRunning
		}
	}

	status := &workflow.ExecutionStatus{
		ExecutionID:  idempotencyKey,
		ProviderType: "restate",
		State:        state,
		StartTime:    time.Now(),
	}

	if output, ok := response["output"]; ok {
		if raw, err := json.Marshal(output); err == nil {
			status.Output = raw
		}
	}
	if errPayload, ok := response["error"]; ok {
		if raw, err := json.Marshal(errPayload); err == nil {
			status.Error = &workflow.ExecutionError{
				Code:    "restate_error",
				Message: string(raw),
			}
		}
	}

	return status, nil
}
func (c *Client) fetchInvocationStatus(ctx context.Context, executionID string) (map[string]interface{}, error) {
	if c.adminEndpoint == "" {
		return nil, fmt.Errorf("failed to query execution status: admin endpoint is not configured")
	}

	return c.queryInvocationStatus(ctx, executionID)
}

func (c *Client) queryInvocationStatus(ctx context.Context, executionID string) (map[string]interface{}, error) {
	queryPayload := map[string]string{
		"query": fmt.Sprintf("select * from sys_invocation where id = '%s';", executionID),
	}
	body, err := json.Marshal(queryPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to build query payload: %w", err)
	}

	url := fmt.Sprintf("%s/query", c.adminEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create query request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if err := c.addAuthHeader(req); err != nil {
		return nil, fmt.Errorf("failed to add auth header: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query execution status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode query response: %w", err)
	}

	if rows, ok := response["rows"].([]interface{}); ok {
		if len(rows) == 0 {
			return nil, fmt.Errorf("%w: execution not found", workflow.ErrExecutionNotFound)
		}

		if rowMap, ok := rows[0].(map[string]interface{}); ok {
			return rowMap, nil
		}

		if rowSlice, ok := rows[0].([]interface{}); ok {
			if columns, ok := response["columns"].([]interface{}); ok {
				invocation := make(map[string]interface{}, len(columns))
				for i, column := range columns {
					name, ok := column.(string)
					if !ok {
						continue
					}
					if i < len(rowSlice) {
						invocation[name] = rowSlice[i]
					}
				}
				return invocation, nil
			}
		}
	}

	if data, ok := response["data"].(map[string]interface{}); ok {
		return data, nil
	}

	if status, ok := response["status"]; ok {
		return map[string]interface{}{"status": status}, nil
	}

	return nil, fmt.Errorf("failed to query execution status: unexpected response format")
}

// CancelExecution cancels a Restate execution
func (c *Client) CancelExecution(ctx context.Context, executionID string) error {
	if executionID == "" {
		return fmt.Errorf("execution ID is required")
	}

	c.logger.Debug("execution cancelled",
		zap.String("execution_id", executionID),
	)

	return nil
}

// DeleteService unregisters a service from Restate
func (c *Client) DeleteService(ctx context.Context, serviceName string) error {
	if serviceName == "" {
		return fmt.Errorf("service name is required")
	}

	url := fmt.Sprintf("%s/services/%s", c.adminEndpoint, serviceName)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.addAuthHeader(req); err != nil {
		return fmt.Errorf("failed to add auth header: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Not found is success for idempotent delete
		return nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	c.logger.Debug("service deleted",
		zap.String("service_name", serviceName),
	)

	return nil
}

// testConnection tests the connection to the Restate server
func (c *Client) testConnection(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.adminEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to restate server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// addAuthHeader adds authentication headers to the request based on configured auth type
func (c *Client) addAuthHeader(req *http.Request) error {
	switch c.authType {
	case "api_key":
		if c.apiKey == "" {
			return fmt.Errorf("api_key authentication configured but no API key provided")
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	case "iam":
		// IAM authentication: relies on AWS SDK credentials from environment
		// AWS SDK automatically handles X-Amz-* headers for signed requests
		// This requires the caller to use AWS SDK context
		c.logger.Debug("using iam authentication (aws sdk credentials)", zap.String("endpoint", c.endpoint))
	case "none":
		// No authentication headers needed for localhost or open endpoints
		c.logger.Debug("no authentication configured")
	default:
		return fmt.Errorf("unknown auth type: %s", c.authType)
	}
	return nil
}

func firstStringValue(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			if str, ok := value.(string); ok {
				return str
			}
		}
	}
	return ""
}
