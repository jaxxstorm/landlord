package main

import (
	"context"
	"fmt"

	"github.com/jaxxstorm/landlord/internal/api/models"
	cliapi "github.com/jaxxstorm/landlord/internal/cli"
	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	var tenantName string
	var config string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tenant",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if tenantName == "" {
				return fmt.Errorf("tenant-name is required")
			}
			if config == "" {
				return fmt.Errorf("config is required")
			}

			client := cliapi.NewClient(cfg.APIURL)
			req := models.CreateTenantRequest{
				Name:  tenantName,
			}
			parsed, err := parseConfigInput(config)
			if err != nil {
				return err
			}
			req.ComputeConfig = parsed
			tenant, err := client.CreateTenant(context.Background(), req)
			if err != nil {
				return err
			}

			cmd.Println(successStyle.Render("Tenant created"))
			cmd.Println(renderTenantDetails(*tenant))
			return nil
		},
	}

	cmd.Flags().StringVar(&tenantName, "tenant-name", "", "Tenant name")
	cmd.Flags().StringVar(&config, "config", "", "Compute config JSON or path to JSON file")

	return cmd
}
