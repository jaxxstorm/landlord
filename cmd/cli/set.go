package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jaxxstorm/landlord/internal/api/models"
	cliapi "github.com/jaxxstorm/landlord/internal/cli"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newSetCommand() *cobra.Command {
	var tenantID string
	var tenantName string
	var config string
	var usePatch bool

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set tenant configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			target := tenantID
			if target == "" {
				target = tenantName
			}
			if target == "" {
				return fmt.Errorf("tenant-id or tenant-name is required")
			}
			if config == "" {
				return fmt.Errorf("config is required")
			}

			req := models.UpdateTenantRequest{}
			if config != "" {
				parsed, err := parseConfigInput(config)
				if err != nil {
					return err
				}
				req.ComputeConfig = parsed
			}

			method := http.MethodPut
			if usePatch {
				method = http.MethodPatch
			}

			client := cliapi.NewClient(cfg.APIURL)
			tenant, err := client.UpdateTenant(context.Background(), target, method, req)
			if err != nil {
				return err
			}

			cmd.Println(successStyle.Render("Tenant updated"))
			cmd.Println(renderTenantDetails(*tenant))
			return nil
		},
	}

	cmd.Flags().StringVar(&tenantID, "tenant-id", "", "Tenant UUID")
	cmd.Flags().StringVar(&tenantName, "tenant-name", "", "Tenant name")
	cmd.Flags().StringVar(&config, "config", "", "Compute config JSON or path to JSON file")
	cmd.Flags().BoolVar(&usePatch, "patch", false, "Use PATCH instead of PUT")

	return cmd
}

func parseConfigInput(value string) (map[string]interface{}, error) {
	if value == "" {
		return nil, nil
	}

	raw := []byte(value)
	sourcePath := ""

	if strings.HasPrefix(value, "file://") {
		path, err := parseFileURI(value)
		if err != nil {
			return nil, err
		}
		sourcePath = path
	} else if info, err := os.Stat(value); err == nil && !info.IsDir() {
		sourcePath = value
	}

	if sourcePath != "" {
		data, err := os.ReadFile(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		raw = data
	}

	ext := strings.ToLower(filepath.Ext(sourcePath))
	switch ext {
	case ".json":
		return parseConfigJSON(raw)
	case ".yaml", ".yml":
		return parseConfigYAML(raw)
	}

	if parsed, err := parseConfigJSON(raw); err == nil {
		return parsed, nil
	} else if parsed, yamlErr := parseConfigYAML(raw); yamlErr == nil {
		return parsed, nil
	} else {
		return nil, fmt.Errorf("parse config input: %v; %v", err, yamlErr)
	}
}

func parseConfigJSON(raw []byte) (map[string]interface{}, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse config JSON: %w", err)
	}
	return parsed, nil
}

func parseConfigYAML(raw []byte) (map[string]interface{}, error) {
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse config YAML: %w", err)
	}
	return parsed, nil
}

func parseFileURI(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse config file URI: %w", err)
	}
	if parsed.Scheme != "file" {
		return "", fmt.Errorf("unsupported config URI scheme: %s", parsed.Scheme)
	}
	path := parsed.Path
	if parsed.Host != "" && parsed.Host != "localhost" {
		// For file:// URLs with relative paths like file://docs/path,
		// the URL parser treats "docs" as the host. Reconstruct the relative path.
		path = parsed.Host + path
	}
	if path == "" {
		path = parsed.Opaque
	}
	if path == "" {
		return "", fmt.Errorf("config file URI missing path")
	}
	unescaped, err := url.PathUnescape(path)
	if err != nil {
		return "", fmt.Errorf("decode config file URI: %w", err)
	}
	if strings.HasPrefix(unescaped, "~") {
		return "", fmt.Errorf("config file URI must use an absolute or relative path, got %s", unescaped)
	}
	return unescaped, nil
}
