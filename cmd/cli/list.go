package main

import (
	"context"

	cliapi "github.com/jaxxstorm/landlord/internal/cli"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	var includeDeleted bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tenants",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client := cliapi.NewClient(cfg.APIURL)
			list, err := client.ListTenants(context.Background(), includeDeleted)
			if err != nil {
				return err
			}

			cmd.Println(renderTenantList(list.Tenants))
			return nil
		},
	}

	cmd.Flags().BoolVar(&includeDeleted, "include-deleted", false, "Include archived tenants")

	return cmd
}
