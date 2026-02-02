package main

import (
	"context"

	cliapi "github.com/jaxxstorm/landlord/internal/cli"
	"github.com/spf13/cobra"
)

func newComputeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute",
		Short: "Show compute config schema for the active provider",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client := cliapi.NewClient(cfg.APIURL)
			resp, err := client.GetComputeConfigDiscovery(context.Background())
			if err != nil {
				return err
			}

			cmd.Println(successStyle.Render("Compute config discovery"))
			cmd.Println(renderComputeConfigDiscovery(*resp))
			return nil
		},
	}

	return cmd
}
