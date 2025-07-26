package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "ssh-cred-manager-cli version",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Aliases:      []string{"v"},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "ssh-cli: %s\n", version)
		},
	}
}
