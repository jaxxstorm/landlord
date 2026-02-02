package main

import (
	"context"
	"fmt"

	cliapi "github.com/jaxxstorm/landlord/internal/cli"
	"github.com/spf13/cobra"
)

func newArchiveCommand() *cobra.Command {
	var tenantID string
	var tenantName string

	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive a tenant",
		RunE: func(cmd *cobra.Command, _ []string) error {
			target := tenantID
			if target == "" {
				target = tenantName
			}
			if target == "" {
				return fmt.Errorf("tenant-id or tenant-name is required")
			}

			client := cliapi.NewClient(cfg.APIURL)
			tenant, err := client.ArchiveTenant(context.Background(), target)
			if err != nil {
				return err
			}

			cmd.Println(successStyle.Render("Tenant archival requested"))
			if tenant != nil {
				cmd.Println(renderTenantDetails(*tenant))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&tenantID, "tenant-id", "", "Tenant UUID")
	cmd.Flags().StringVar(&tenantName, "tenant-name", "", "Tenant name")

	return cmd
}
