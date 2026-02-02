package main

import (
	"context"
	"fmt"

	cliapi "github.com/jaxxstorm/landlord/internal/cli"
	"github.com/spf13/cobra"
)

func newGetCommand() *cobra.Command {
	var tenantID string
	var tenantName string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a tenant",
		RunE: func(cmd *cobra.Command, _ []string) error {
			target := tenantID
			if target == "" {
				target = tenantName
			}
			if target == "" {
				return fmt.Errorf("tenant-id or tenant-name is required")
			}

			client := cliapi.NewClient(cfg.APIURL)
			tenant, err := client.GetTenant(context.Background(), target)
			if err != nil {
				return err
			}

			cmd.Println(headerStyle.Render("Tenant details"))
			cmd.Println(renderTenantDetails(*tenant))
			return nil
		},
	}

	cmd.Flags().StringVar(&tenantID, "tenant-id", "", "Tenant UUID")
	cmd.Flags().StringVar(&tenantName, "tenant-name", "", "Tenant name")

	return cmd
}
