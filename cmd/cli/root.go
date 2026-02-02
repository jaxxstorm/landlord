package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "landlord-cli",
		Short: "CLI for interacting with the Landlord API",
		Long:  "A command-line tool for tenant lifecycle operations via the Landlord API.",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return loadCLIConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().String("config", "", "Config file path")
	cmd.PersistentFlags().String("api-url", "http://localhost:8081", "Landlord API base URL (versioned paths are appended if missing)")

	if err := bindCLIFlags(cmd); err != nil {
		cmd.PrintErrln(fmt.Sprintf("failed to bind flags: %v", err))
	}

	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newComputeCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newArchiveCommand())
	cmd.AddCommand(newDeleteCommand())

	return cmd
}
