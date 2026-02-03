package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"strings"

	"github.com/google/uuid"
	"github.com/jaxxstorm/landlord/internal/api/models"
	"github.com/jaxxstorm/landlord/internal/apiversion"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	baseURL = apiversion.NormalizeBaseURL(baseURL)
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) CreateTenant(ctx context.Context, req models.CreateTenantRequest) (*models.TenantResponse, error) {
	url := fmt.Sprintf("%s/tenants", c.baseURL)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := handleErrorResponse(resp); err != nil {
		return nil, err
	}

	var tenant models.TenantResponse
	if err := json.NewDecoder(resp.Body).Decode(&tenant); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &tenant, nil
}

func (c *Client) ListTenants(ctx context.Context, includeDeleted bool) (*models.ListTenantsResponse, error) {
	url := fmt.Sprintf("%s/tenants", c.baseURL)
	if includeDeleted {
		url = fmt.Sprintf("%s?include_deleted=true", url)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := handleErrorResponse(resp); err != nil {
		return nil, err
	}

	var list models.ListTenantsResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &list, nil
}

func (c *Client) DeleteTenant(ctx context.Context, tenantID string) (*models.TenantResponse, error) {
	id, err := c.resolveTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/tenants/%s", c.baseURL, id)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := handleErrorResponse(resp); err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	var tenant models.TenantResponse
	if err := json.NewDecoder(resp.Body).Decode(&tenant); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &tenant, nil
}

func (c *Client) ArchiveTenant(ctx context.Context, tenantID string) (*models.TenantResponse, error) {
	id, err := c.resolveTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/tenants/%s/archive", c.baseURL, id)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := handleErrorResponse(resp); err != nil {
		return nil, err
	}

	var tenant models.TenantResponse
	if err := json.NewDecoder(resp.Body).Decode(&tenant); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &tenant, nil
}

func (c *Client) GetTenant(ctx context.Context, tenantID string) (*models.TenantResponse, error) {
	id, err := c.resolveTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/tenants/%s", c.baseURL, id)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := handleErrorResponse(resp); err != nil {
		return nil, err
	}

	var tenant models.TenantResponse
	if err := json.NewDecoder(resp.Body).Decode(&tenant); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &tenant, nil
}

func (c *Client) UpdateTenant(ctx context.Context, tenantID string, method string, req models.UpdateTenantRequest) (*models.TenantResponse, error) {
	id, err := c.resolveTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/tenants/%s", c.baseURL, id)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := handleErrorResponse(resp); err != nil {
		return nil, err
	}

	var tenant models.TenantResponse
	if err := json.NewDecoder(resp.Body).Decode(&tenant); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &tenant, nil
}

func (c *Client) GetComputeConfigDiscovery(ctx context.Context, provider string) (*models.ComputeConfigDiscoveryResponse, error) {
	url := fmt.Sprintf("%s/compute/config", c.baseURL)
	if provider != "" {
		url = fmt.Sprintf("%s?provider=%s", url, provider)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := handleErrorResponse(resp); err != nil {
		return nil, err
	}

	var discovery models.ComputeConfigDiscoveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &discovery, nil
}

func (c *Client) resolveTenantID(ctx context.Context, tenantID string) (string, error) {
	if _, err := uuid.Parse(tenantID); err == nil {
		return tenantID, nil
	}

	list, err := c.ListTenants(ctx, true)
	if err != nil {
		return "", err
	}

	for _, tenant := range list.Tenants {
		if tenant.Name == tenantID {
			return tenant.ID, nil
		}
	}

	return "", fmt.Errorf("tenant not found: %s", tenantID)
}

func handleErrorResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		return fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	var apiErr models.ErrorResponse
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return fmt.Errorf("api error: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if apiErr.Error != "" {
		return fmt.Errorf("api error: %s", apiErr.Error)
	}

	return fmt.Errorf("api error: status %d", resp.StatusCode)
}
