package docker

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestProvider_ValidateConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)
	provider, err := New(&Config{}, logger)
	if err != nil {
		t.Skipf("skipping docker provider validation: %v", err)
	}

	tests := []struct {
		name    string
		config  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty config is valid",
			config:  "",
			wantErr: false,
		},
		{
			name:    "valid config with env only",
			config:  `{"env": {"PORT": "8080", "DEBUG": "true"}}`,
			wantErr: false,
		},
		{
			name:    "valid config with volumes",
			config:  `{"volumes": ["/data:/app/data", "/logs:/app/logs:ro"]}`,
			wantErr: false,
		},
		{
			name:    "valid config with network_mode bridge",
			config:  `{"network_mode": "bridge"}`,
			wantErr: false,
		},
		{
			name:    "valid config with network_mode host",
			config:  `{"network_mode": "host"}`,
			wantErr: false,
		},
		{
			name:    "valid config with container network",
			config:  `{"network_mode": "container:mycontainer"}`,
			wantErr: false,
		},
		{
			name:    "valid config with ports",
			config:  `{"ports": [{"container_port": 8080, "host_port": 8080, "protocol": "tcp"}]}`,
			wantErr: false,
		},
		{
			name:    "valid config with restart policy",
			config:  `{"restart_policy": "always"}`,
			wantErr: false,
		},
		{
			name:    "valid full config",
			config:  `{"env": {"PORT": "8080"}, "volumes": ["/data:/app/data"], "network_mode": "bridge", "ports": [{"container_port": 8080}], "restart_policy": "unless-stopped"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			config:  `{invalid json}`,
			wantErr: true,
			errMsg:  "invalid JSON structure",
		},
		{
			name:    "invalid volume format - missing colon",
			config:  `{"volumes": ["/data"]}`,
			wantErr: true,
			errMsg:  "invalid format",
		},
		{
			name:    "invalid volume format - relative container path",
			config:  `{"volumes": ["/data:relative/path"]}`,
			wantErr: true,
			errMsg:  "container path must be absolute",
		},
		{
			name:    "invalid network_mode",
			config:  `{"network_mode": "invalid"}`,
			wantErr: true,
			errMsg:  "invalid value",
		},
		{
			name:    "invalid port - container port too low",
			config:  `{"ports": [{"container_port": 0}]}`,
			wantErr: true,
			errMsg:  "must be between 1 and 65535",
		},
		{
			name:    "invalid port - container port too high",
			config:  `{"ports": [{"container_port": 70000}]}`,
			wantErr: true,
			errMsg:  "must be between 1 and 65535",
		},
		{
			name:    "invalid port - host port too high",
			config:  `{"ports": [{"container_port": 8080, "host_port": 70000}]}`,
			wantErr: true,
			errMsg:  "must be between 1 and 65535",
		},
		{
			name:    "invalid port - invalid protocol",
			config:  `{"ports": [{"container_port": 8080, "protocol": "http"}]}`,
			wantErr: true,
			errMsg:  "must be 'tcp' or 'udp'",
		},
		{
			name:    "invalid restart policy",
			config:  `{"restart_policy": "sometimes"}`,
			wantErr: true,
			errMsg:  "invalid value",
		},
		{
			name:    "invalid env - contains equals",
			config:  `{"env": {"KEY=VALUE": "test"}}`,
			wantErr: true,
			errMsg:  "cannot contain '='",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configJSON json.RawMessage
			if tt.config != "" {
				configJSON = json.RawMessage(tt.config)
			}

			err := provider.ValidateConfig(configJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
