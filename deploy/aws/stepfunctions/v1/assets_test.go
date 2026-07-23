package stepfunctionsv1

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLifecycleASLDefinition(t *testing.T) {
	raw, err := os.ReadFile("lifecycle.asl.json")
	require.NoError(t, err)

	var definition map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &definition))
	require.Equal(t, "InvokeLifecycle", definition["StartAt"])

	states, ok := definition["States"].(map[string]interface{})
	require.True(t, ok)
	invoke, ok := states["InvokeLifecycle"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "arn:aws:states:::lambda:invoke", invoke["Resource"])
	parameters, ok := invoke["Parameters"].(map[string]interface{})
	require.True(t, ok)
	payload, ok := parameters["Payload"].(map[string]interface{})
	require.True(t, ok)
	require.Contains(t, payload, "workflow_execution_id.$")
	require.NotEmpty(t, invoke["Retry"])
	require.NotEmpty(t, invoke["Catch"])
}

func TestDeploymentTemplateStructure(t *testing.T) {
	raw, err := os.ReadFile("template.yaml")
	require.NoError(t, err)

	var template map[string]interface{}
	require.NoError(t, yaml.Unmarshal(raw, &template))
	resources, ok := template["Resources"].(map[string]interface{})
	require.True(t, ok)
	for _, name := range []string{"LifecycleLambda", "LifecycleStateMachine", "LifecycleLambdaRole", "LifecycleStateMachineRole"} {
		require.Contains(t, resources, name)
	}
	parameters, ok := template["Parameters"].(map[string]interface{})
	require.True(t, ok)
	require.Contains(t, parameters, "LambdaCodeBucket")
	require.Contains(t, parameters, "StateMachineDefinitionBucket")
}
