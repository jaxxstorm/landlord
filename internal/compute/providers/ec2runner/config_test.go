package ec2runner

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigSchemaAndValidation(t *testing.T) {
	provider := &Provider{}
	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal(provider.ConfigSchema(), &schema))
	require.Equal(t, []interface{}{"ami", "user_data"}, schema["required"])

	require.NoError(t, provider.ValidateConfig(json.RawMessage(`{"ami":"ami-123","user_data":"#!/bin/bash"}`)))
	require.ErrorContains(t, provider.ValidateConfig(json.RawMessage(`{"ami":"ami-123"}`)), "user_data is required")
	require.ErrorContains(t, provider.ValidateConfig(json.RawMessage(`{"ami":"ami-123","user_data":"x","instance_type":"t3.micro"}`)), "unsupported EC2 runner config field")
	require.Nil(t, provider.ConfigDefaults())
}
