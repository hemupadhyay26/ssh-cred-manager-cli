package tui

import (
	model "github.com/hemupadhyay26/ssh-cred-manager-cli/tui"
	"github.com/spf13/cobra"
)

func NewTuiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tui",
		Short:   "Launch the SSH Credential Manager TUI",
		Long:    `A terminal user interface for managing SSH credentials interactively.`,
		Aliases: []string{"t", "tu"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return model.Run() // ✅ use Run not main
		},
	}
	return cmd
}
