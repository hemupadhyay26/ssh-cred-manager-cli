package cmd

import (
	"fmt"

	"github.com/hemupadhyay26/ssh-cred-manager-cli/cmd/ssh"
	"github.com/spf13/cobra"
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh-cli",
		Short: "cli tool help you to ssh management",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newVersionCmd(version)) // version subcommand
	cmd.AddCommand(ssh.NewSSHCmd())
	// Register the man command
	cmd.AddCommand(NewManCmd().Cmd)

	return cmd
}

// Execute invokes the command.
func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
