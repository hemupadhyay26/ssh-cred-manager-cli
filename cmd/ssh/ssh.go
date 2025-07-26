package ssh

import (
	"github.com/spf13/cobra"
)

func NewSSHCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ssh",
		Short:   "Manage and connect to SSH servers",
		Long:    `A suite of commands to save, list, delete, and connect to SSH servers.`,
		Aliases: []string{"s", "ss"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewSaveCmd())
	cmd.AddCommand(NewSaveWizardCmd())
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewDeleteCmd())
	cmd.AddCommand(NewConnectCmd())
	cmd.AddCommand(NewUpdateCmd())

	return cmd
}
