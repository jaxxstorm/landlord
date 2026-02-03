package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/jaxxstorm/landlord/internal/api/models"
	cliapi "github.com/jaxxstorm/landlord/internal/cli"
	"github.com/spf13/cobra"
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
	if info, err := os.Stat(value); err == nil && !info.IsDir() {
		data, err := os.ReadFile(value)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		raw = data
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse config JSON: %w", err)
	}

	return parsed, nil
}
