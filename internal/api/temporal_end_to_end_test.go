package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func requireTemporalE2E(t *testing.T) string {
	t.Helper()
	if os.Getenv("TEMPORAL_E2E_TEST") == "" {
		t.Skip("skipping Temporal e2e test: TEMPORAL_E2E_TEST not set")
	}
	baseURL := os.Getenv("LANDLORD_E2E_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}
	return baseURL
}

func TestTemporalE2ECreateUpdateDeleteTenant(t *testing.T) {
	baseURL := requireTemporalE2E(t)
	client := &http.Client{Timeout: 15 * time.Second}

	name := fmt.Sprintf("temporal-e2e-%d", time.Now().UnixNano())
	createBody := map[string]interface{}{
		"name": name,
		"compute_config": map[string]interface{}{
			"image": "nginx:latest",
		},
	}

	created := map[string]interface{}{}
	status := doJSONRequest(t, client, http.MethodPost, baseURL+"/v1/tenants", createBody, &created)
	require.Equal(t, http.StatusCreated, status)

	id, ok := created["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, id)

	updateBody := map[string]interface{}{
		"compute_config": map[string]interface{}{
			"image": "nginx:1.27",
		},
	}
	status = doJSONRequest(t, client, http.MethodPut, baseURL+"/v1/tenants/"+id, updateBody, nil)
	require.Contains(t, []int{http.StatusOK, http.StatusAccepted}, status)

	status = doJSONRequest(t, client, http.MethodDelete, baseURL+"/v1/tenants/"+id, nil, nil)
	require.Equal(t, http.StatusAccepted, status)
}

func TestTemporalE2ECancellationPath(t *testing.T) {
	baseURL := requireTemporalE2E(t)
	client := &http.Client{Timeout: 15 * time.Second}

	name := fmt.Sprintf("temporal-cancel-%d", time.Now().UnixNano())
	createBody := map[string]interface{}{
		"name": name,
		"compute_config": map[string]interface{}{
			"image": "nginx:latest",
		},
	}

	created := map[string]interface{}{}
	status := doJSONRequest(t, client, http.MethodPost, baseURL+"/v1/tenants", createBody, &created)
	require.Equal(t, http.StatusCreated, status)
	id := created["id"].(string)

	status = doJSONRequest(t, client, http.MethodDelete, baseURL+"/v1/tenants/"+id, nil, nil)
	require.Equal(t, http.StatusAccepted, status)

	// Poll for terminal deletion/archive progression.
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		current := map[string]interface{}{}
		status = doJSONRequest(t, client, http.MethodGet, baseURL+"/v1/tenants/"+id, nil, &current)
		if status == http.StatusNotFound || status == http.StatusGone {
			return
		}
		if status == http.StatusOK {
			if rawStatus, ok := current["status"].(string); ok {
				if rawStatus == "archiving" || rawStatus == "deleting" || rawStatus == "archived" {
					return
				}
			}
			if _, hasSubState := current["workflow_sub_state"]; hasSubState {
				return
			}
			if _, hasExecutionID := current["workflow_execution_id"]; hasExecutionID {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	t.Fatalf("tenant %s did not reach archived/deleting/notfound state within timeout", id)
}

func doJSONRequest(t *testing.T, client *http.Client, method, url string, body map[string]interface{}, out interface{}) int {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode
}
