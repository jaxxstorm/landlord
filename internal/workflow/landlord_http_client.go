package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// HTTPLandlordClient fetches tenant data from the landlord HTTP API.
type HTTPLandlordClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewHTTPLandlordClient creates a new HTTP client for the landlord API.
func NewHTTPLandlordClient(baseURL string, logger *zap.Logger) *HTTPLandlordClient {
	return &HTTPLandlordClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger.With(zap.String("component", "landlord-http-client")),
	}
}

// GetTenant retrieves a tenant by UUID from the landlord API.
func (c *HTTPLandlordClient) GetTenant(ctx context.Context, tenantUUID string) (*LandlordTenant, error) {
	if tenantUUID == "" {
		return nil, fmt.Errorf("tenant UUID is required")
	}

	url := fmt.Sprintf("%s/api/tenants/%s", c.baseURL, tenantUUID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request tenant: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tenant LandlordTenant
	if err := json.NewDecoder(resp.Body).Decode(&tenant); err != nil {
		return nil, fmt.Errorf("decode tenant: %w", err)
	}

	return &tenant, nil
}
