package stepfunctions

import (
	"strings"
	"testing"

	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutionNameIsStableForSameLifecycleRevision(t *testing.T) {
	request := &workflow.ProvisionRequest{
		TenantUUID: "6f959cd1-b087-4b21-a0ed-02b4fa011271",
		Operation:  "update",
		Metadata:   map[string]string{"config_hash": "desired-state-revision"},
	}

	first, err := executionName(request)
	require.NoError(t, err)
	second, err := executionName(request)
	require.NoError(t, err)

	assert.Equal(t, first, second)
	assert.LessOrEqual(t, len(first), maxExecutionNameLength)
	assert.Regexp(t, `^[a-z0-9_-]+$`, first)
}

func TestExecutionNameChangesWithDesiredStateRevision(t *testing.T) {
	request := &workflow.ProvisionRequest{
		TenantUUID: "6f959cd1-b087-4b21-a0ed-02b4fa011271",
		Operation:  "provision",
		Metadata:   map[string]string{"config_hash": "revision-one"},
	}

	first, err := executionName(request)
	require.NoError(t, err)
	request.Metadata["config_hash"] = "revision-two"
	second, err := executionName(request)
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
}

func TestExecutionNameNormalizesAndBoundsInputs(t *testing.T) {
	request := &workflow.ProvisionRequest{
		TenantID:  strings.Repeat("tenant name/", 20),
		Operation: "Update Tenant!",
		Metadata:  map[string]string{"config_hash": "revision"},
	}

	name, err := executionName(request)
	require.NoError(t, err)

	assert.LessOrEqual(t, len(name), maxExecutionNameLength)
	assert.Regexp(t, `^[a-z0-9_-]+$`, name)
}

func TestExecutionNameRequiresTenantIdentifier(t *testing.T) {
	_, err := executionName(&workflow.ProvisionRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant identifier")
}
