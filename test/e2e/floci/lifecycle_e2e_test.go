//go:build e2e

package floci_test

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/jaxxstorm/landlord/internal/cloud/awsconfig"
	"github.com/jaxxstorm/landlord/internal/compute"
)

const (
	lifecycleTimeout = 2 * time.Minute
	pollInterval     = time.Second
)

type tenantResponse struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
}

func TestEC2TenantLifecycle(t *testing.T) {
	accountID := randomAccountID(t)
	stackID := "floci-e2e-" + accountID
	stackEnv := []string{
		"LANDLORD_FLOCI_E2E_ID=" + stackID,
		"LANDLORD_FLOCI_ACCOUNT_ID=" + accountID,
	}

	output, err := runStack(stackEnv, "up")
	t.Cleanup(func() {
		if t.Failed() {
			if logs, logsErr := runStack(stackEnv, "logs"); logsErr != nil {
				t.Logf("collect stack logs: %v\n%s", logsErr, logs)
			} else {
				t.Logf("Floci stack logs:\n%s", logs)
			}
		}
		if downOutput, downErr := runStack(stackEnv, "down"); downErr != nil {
			t.Errorf("stop Floci stack: %v\n%s", downErr, downOutput)
		}
	})
	if err != nil {
		t.Fatalf("start Floci stack: %v\n%s", err, output)
	}

	stackConfig, err := readStackConfig(filepath.Join("..", "..", "..", "dist", "floci-e2e", stackID, "stack.env"))
	if err != nil {
		t.Fatalf("read stack configuration: %v", err)
	}
	apiURL := stackConfig["LANDLORD_API_URL"]
	flociEndpoint := stackConfig["FLOCI_ENDPOINT"]
	if apiURL == "" || flociEndpoint == "" {
		t.Fatalf("stack configuration does not include API and Floci endpoints: %#v", stackConfig)
	}

	tenant := createTenant(t, apiURL, "floci-e2e-"+accountID)
	tenant = waitForTenant(t, apiURL, tenant.ID, "ready", "failed")
	if tenant.Status != "ready" {
		t.Fatalf("tenant provisioning ended in %q: %s", tenant.Status, tenant.StatusMessage)
	}

	ec2Client := newEC2Client(t, flociEndpoint, accountID)
	instances := waitForTaggedInstances(t, ec2Client, tenant.ID, true)
	if len(instances) != 1 {
		t.Fatalf("expected one active tagged EC2 runner, found %d", len(instances))
	}

	deleteTenant(t, apiURL, tenant.ID)
	waitForTenantDeletion(t, apiURL, tenant.ID)
	waitForTaggedInstances(t, ec2Client, tenant.ID, false)
}

func randomAccountID(t *testing.T) string {
	t.Helper()
	random, err := cryptorand.Int(cryptorand.Reader, big.NewInt(900_000_000_000))
	if err != nil {
		t.Fatalf("generate Floci account ID: %v", err)
	}
	return fmt.Sprintf("%012d", random.Int64()+100_000_000_000)
}

func runStack(env []string, args ...string) (string, error) {
	command := exec.Command("./stack.sh", args...)
	command.Env = append(os.Environ(), env...)
	output, err := command.CombinedOutput()
	return string(output), err
}

func readStackConfig(path string) (map[string]string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := make(map[string]string)
	for _, line := range strings.Split(string(contents), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok {
			config[key] = value
		}
	}
	return config, nil
}

func createTenant(t *testing.T, apiURL, name string) tenantResponse {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"name": name,
		"compute_config": map[string]string{
			"ami":       "ami-alpine",
			"user_data": "#!/bin/sh\necho landlord-floci-e2e",
		},
	})
	if err != nil {
		t.Fatalf("marshal tenant request: %v", err)
	}

	response := tenantResponse{}
	requestTenant(t, http.MethodPost, apiURL+"/v1/tenants", bytes.NewReader(body), http.StatusCreated, &response)
	if response.ID == "" {
		t.Fatal("create tenant response did not include an ID")
	}
	return response
}

func deleteTenant(t *testing.T, apiURL, tenantID string) {
	t.Helper()
	requestTenant(t, http.MethodDelete, apiURL+"/v1/tenants/"+url.PathEscape(tenantID), nil, http.StatusAccepted, nil)
}

func waitForTenant(t *testing.T, apiURL, tenantID string, statuses ...string) tenantResponse {
	t.Helper()
	deadline := time.Now().Add(lifecycleTimeout)
	for {
		response := tenantResponse{}
		status, body, err := tenantRequest(http.MethodGet, apiURL+"/v1/tenants/"+url.PathEscape(tenantID), nil, &response)
		if err == nil && status == http.StatusOK {
			for _, expected := range statuses {
				if response.Status == expected {
					return response
				}
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("tenant %s did not reach %v; last response status=%d body=%s error=%v", tenantID, statuses, status, body, err)
		}
		time.Sleep(pollInterval)
	}
}

func waitForTenantDeletion(t *testing.T, apiURL, tenantID string) {
	t.Helper()
	deadline := time.Now().Add(lifecycleTimeout)
	for {
		response := tenantResponse{}
		status, body, err := tenantRequest(http.MethodGet, apiURL+"/v1/tenants/"+url.PathEscape(tenantID), nil, &response)
		if err == nil && status == http.StatusNotFound {
			return
		}
		if err == nil && status == http.StatusOK && response.Status == "failed" {
			t.Fatalf("tenant deletion failed: %s", response.StatusMessage)
		}
		if time.Now().After(deadline) {
			t.Fatalf("tenant %s was not deleted; last response status=%d body=%s error=%v", tenantID, status, body, err)
		}
		time.Sleep(pollInterval)
	}
}

func requestTenant(t *testing.T, method, requestURL string, body io.Reader, expectedStatus int, response any) {
	t.Helper()
	status, responseBody, err := tenantRequest(method, requestURL, body, response)
	if err != nil {
		t.Fatalf("%s %s: %v", method, requestURL, err)
	}
	if status != expectedStatus {
		t.Fatalf("%s %s returned %d, expected %d: %s", method, requestURL, status, expectedStatus, responseBody)
	}
}

func tenantRequest(method, requestURL string, body io.Reader, response any) (int, string, error) {
	request, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return 0, "", err
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 15 * time.Second}
	result, err := client.Do(request)
	if err != nil {
		return 0, "", err
	}
	defer result.Body.Close()
	responseBody, err := io.ReadAll(result.Body)
	if err != nil {
		return result.StatusCode, "", err
	}
	if response != nil && len(responseBody) > 0 && result.StatusCode < http.StatusBadRequest {
		if err := json.Unmarshal(responseBody, response); err != nil {
			return result.StatusCode, string(responseBody), err
		}
	}
	return result.StatusCode, string(responseBody), nil
}

func newEC2Client(t *testing.T, endpoint, accountID string) *ec2.Client {
	t.Helper()
	config, err := awsconfig.Load(context.Background(), awsconfig.Options{
		Region:              "us-east-1",
		EndpointURL:         endpoint,
		CredentialsProvider: credentials.NewStaticCredentialsProvider(accountID, "test", ""),
	})
	if err != nil {
		t.Fatalf("load Floci AWS configuration: %v", err)
	}
	return ec2.NewFromConfig(config)
}

func waitForTaggedInstances(t *testing.T, client *ec2.Client, tenantID string, shouldExist bool) []ec2types.Instance {
	t.Helper()
	deadline := time.Now().Add(lifecycleTimeout)
	for {
		instances, err := taggedInstances(context.Background(), client, tenantID)
		if err == nil && (len(instances) > 0) == shouldExist {
			return instances
		}
		if time.Now().After(deadline) {
			t.Fatalf("tagged EC2 runner existence=%t for tenant %s was not observed; instances=%d error=%v", shouldExist, tenantID, len(instances), err)
		}
		time.Sleep(pollInterval)
	}
}

func taggedInstances(ctx context.Context, client *ec2.Client, tenantID string) ([]ec2types.Instance, error) {
	result, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{Filters: []ec2types.Filter{
		{Name: aws.String("tag:" + compute.MetadataOwnerKey), Values: []string{compute.MetadataOwnerValue}},
		{Name: aws.String("tag:" + compute.MetadataTenantIDKey), Values: []string{tenantID}},
		{Name: aws.String("tag:" + compute.MetadataProviderKey), Values: []string{"ec2-runner"}},
	}})
	if err != nil {
		return nil, err
	}
	instances := make([]ec2types.Instance, 0)
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if instance.State == nil || instance.State.Name != ec2types.InstanceStateNameTerminated {
				instances = append(instances, instance)
			}
		}
	}
	return instances, nil
}
