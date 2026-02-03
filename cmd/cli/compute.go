package main

import (
	"context"
	"fmt"

	cliapi "github.com/jaxxstorm/landlord/internal/cli"
	"github.com/spf13/cobra"
)

func newComputeCommand() *cobra.Command {
	var provider string

	cmd := &cobra.Command{
		Use:   "compute",
		Short: "Show compute config schema for a provider",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if provider == "" {
				return fmt.Errorf("--provider is required")
			}
			client := cliapi.NewClient(cfg.APIURL)
			resp, err := client.GetComputeConfigDiscovery(context.Background(), provider)
			if err != nil {
				return err
			}

			cmd.Println(successStyle.Render("Compute config discovery"))
			cmd.Println(renderComputeConfigDiscovery(*resp))
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "Compute provider identifier")
	_ = cmd.MarkFlagRequired("provider")

	return cmd
}
