package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewViperInstance(t *testing.T) {
	v := NewViperInstance()

	assert.NotNil(t, v)
	assert.Equal(t, "localhost", v.GetString("database.host"))
	assert.Equal(t, 5432, v.GetInt("database.port"))
	assert.Equal(t, "0.0.0.0", v.GetString("http.host"))
	assert.Equal(t, 8080, v.GetInt("http.port"))
	assert.Equal(t, "info", v.GetString("log.level"))
	assert.Equal(t, "development", v.GetString("log.format"))
}

func TestBindEnvironmentVariables(t *testing.T) {
	v := NewViperInstance()

	err := BindEnvironmentVariables(v)
	require.NoError(t, err)

	// Set environment variables
	t.Setenv("DB_HOST", "testhost")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "testuser")
	t.Setenv("DB_PASSWORD", "testpass")
	t.Setenv("LOG_LEVEL", "debug")

	// Create new instance and bind to pick up env vars
	v2 := NewViperInstance()
	err = BindEnvironmentVariables(v2)
	require.NoError(t, err)

	assert.Equal(t, "testhost", v2.GetString("database.host"))
	assert.Equal(t, "debug", v2.GetString("log.level"))
}

func TestFindConfigFile_ExplicitPath(t *testing.T) {
	// Create a temporary config file
	tempFile, err := os.CreateTemp("", "config*.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Should find the explicit path
	found, err := FindConfigFile(tempFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, tempFile.Name(), found)
}

func TestFindConfigFile_ExplicitPathNotFound(t *testing.T) {
	// Try to find a file that doesn't exist
	_, err := FindConfigFile("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFindConfigFile_EnvironmentVariable(t *testing.T) {
	// Create a temporary config file
	tempFile, err := os.CreateTemp("", "config*.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Set environment variable
	t.Setenv("LANDLORD_CONFIG", tempFile.Name())

	// Should find via environment variable
	found, err := FindConfigFile("")
	assert.NoError(t, err)
	assert.Equal(t, tempFile.Name(), found)
}

func TestFindConfigFile_CurrentDirectory(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create config.yaml in the temp directory
	configPath := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configPath, []byte("test: value"), 0644)
	require.NoError(t, err)

	// Change to temp directory temporarily
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	// Should find in current directory
	found, err := FindConfigFile("")
	assert.NoError(t, err)
	assert.NotEmpty(t, found)
	assert.Contains(t, found, "config.yaml")
}

func TestFindConfigFile_NotFound(t *testing.T) {
	// In a temp directory with no config files
	tempDir, err := os.MkdirTemp("", "config_test_empty")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	// Unset environment variable
	t.Setenv("LANDLORD_CONFIG", "")

	// Should return empty string (not an error - config is optional)
	found, err := FindConfigFile("")
	assert.NoError(t, err)
	assert.Empty(t, found)
}

func TestLoadConfigFile_YAML(t *testing.T) {
	// Create a temporary YAML config file
	tempFile, err := os.CreateTemp("", "config*.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	configContent := `database:
  host: yamlhost
  port: 5433
  user: yamluser
log:
  level: debug`

	err = os.WriteFile(tempFile.Name(), []byte(configContent), 0644)
	require.NoError(t, err)

	v := NewViperInstance()
	err = LoadConfigFile(v, tempFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "yamlhost", v.GetString("database.host"))
	assert.Equal(t, 5433, v.GetInt("database.port"))
	assert.Equal(t, "debug", v.GetString("log.level"))
}

func TestLoadConfigFile_JSON(t *testing.T) {
	// Create a temporary JSON config file
	tempFile, err := os.CreateTemp("", "config*.json")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	configContent := `{
  "database": {
    "host": "jsonhost",
    "port": 5434,
    "user": "jsonuser"
  },
  "log": {
    "level": "warn"
  }
}`

	err = os.WriteFile(tempFile.Name(), []byte(configContent), 0644)
	require.NoError(t, err)

	v := NewViperInstance()
	err = LoadConfigFile(v, tempFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "jsonhost", v.GetString("database.host"))
	assert.Equal(t, 5434, v.GetInt("database.port"))
	assert.Equal(t, "warn", v.GetString("log.level"))
}

func TestLoadConfigFile_InvalidYAML(t *testing.T) {
	tempFile, err := os.CreateTemp("", "config*.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write invalid YAML
	err = os.WriteFile(tempFile.Name(), []byte("invalid: yaml: content: ["), 0644)
	require.NoError(t, err)

	v := NewViperInstance()
	err = LoadConfigFile(v, tempFile.Name())
	assert.Error(t, err)
}

func TestLoadConfigFile_UnsupportedExtension(t *testing.T) {
	tempFile, err := os.CreateTemp("", "config*.toml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	v := NewViperInstance()
	err = LoadConfigFile(v, tempFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestLoadFromViper_Valid(t *testing.T) {
	v := NewViperInstance()
	setComputeDefaults(v)

	// Set some values
	v.Set("database.host", "testhost")
	v.Set("database.port", 5433)
	v.Set("database.user", "testuser")
	v.Set("database.password", "testpass")
	v.Set("database.database", "testdb")
	v.Set("http.port", 8081)

	cfg, err := LoadFromViper(v)
	require.NoError(t, err)

	assert.NotNil(t, cfg)
	assert.Equal(t, "testhost", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testdb", cfg.Database.Database)
	assert.Equal(t, 8081, cfg.HTTP.Port)
}

func TestLoadFromViper_InvalidConfig(t *testing.T) {
	v := NewViperInstance()
	setComputeDefaults(v)

	// Set invalid values that will fail validation
	v.Set("database.port", 99999) // Invalid port

	_, err := LoadFromViper(v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestLoadFromViper_DefaultValues(t *testing.T) {
	v := NewViperInstance()
	setComputeDefaults(v)

	// Set only required fields
	v.Set("database.user", "user")
	v.Set("database.password", "pass")
	v.Set("database.database", "db")

	cfg, err := LoadFromViper(v)
	require.NoError(t, err)

	// Check defaults are applied
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "prefer", cfg.Database.SSLMode)
	assert.Equal(t, int32(25), cfg.Database.MaxConnections)
	assert.Equal(t, "0.0.0.0", cfg.HTTP.Host)
	assert.Equal(t, 8080, cfg.HTTP.Port)
}

func TestConfigPrecedence_CLIOverridesEnv(t *testing.T) {
	// This test demonstrates CLI flag precedence over env vars
	// The actual precedence is tested through integration tests
	// This is a structural test to ensure the functions exist and work

	v := NewViperInstance()
	err := BindEnvironmentVariables(v)
	require.NoError(t, err)

	// Set env var
	t.Setenv("DB_HOST", "envhost")

	// Create new viper and bind
	v2 := NewViperInstance()
	err = BindEnvironmentVariables(v2)
	require.NoError(t, err)

	// Environment variable should be readable
	assert.Equal(t, "envhost", v2.GetString("database.host"))

	// Simulate CLI override
	v2.Set("database.host", "cliflag")
	assert.Equal(t, "cliflag", v2.GetString("database.host"))
}

func TestConfigDurationParsing(t *testing.T) {
	v := NewViperInstance()
	setComputeDefaults(v)

	v.Set("database.connect_timeout", "5s")
	v.Set("http.shutdown_timeout", "15s")

	cfg, err := LoadFromViper(v)
	require.NoError(t, err)

	assert.Equal(t, 5*time.Second, cfg.Database.ConnectTimeout)
	assert.Equal(t, 15*time.Second, cfg.HTTP.ShutdownTimeout)
}

func TestConfigNestedStructMarshaling(t *testing.T) {
	v := NewViperInstance()
	setComputeDefaults(v)

	// Set nested values
	v.Set("workflow.step_functions.region", "us-east-1")
	v.Set("workflow.step_functions.role_arn", "arn:aws:iam::123456789:role/sfn")

	cfg, err := LoadFromViper(v)
	require.NoError(t, err)

	assert.Equal(t, "us-east-1", cfg.Workflow.StepFunctions.Region)
	assert.Equal(t, "arn:aws:iam::123456789:role/sfn", cfg.Workflow.StepFunctions.RoleARN)
}

func setComputeDefaults(v *viper.Viper) {
	v.Set("compute.defaults.docker", map[string]interface{}{
		"image": "nginx:latest",
	})
	v.Set("compute.defaults.ecs", map[string]interface{}{
		"cluster_arn":         "arn:aws:ecs:us-east-1:123456789012:cluster/test",
		"task_definition_arn": "arn:aws:ecs:us-east-1:123456789012:task-definition/test:1",
	})
}
