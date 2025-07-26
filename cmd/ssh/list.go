package ssh

import (
	"fmt"

	"github.com/hemupadhyay26/ssh-cred-manager-cli/internal/credential"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved SSH credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := credential.NewCredentialStore()
			if err != nil {
				return fmt.Errorf("failed to initialize credential store: %w", err)
			}

			credentials := store.ListCredentials()
			if len(credentials) == 0 {
				fmt.Println("No SSH credentials found")
				return nil
			}

			fmt.Println("Saved SSH credentials:")
			fmt.Println("---------------------")
			for _, cred := range credentials {
				fmt.Printf("Name: %s\n", cred.Name)
				fmt.Printf("Host: %s:%d\n", cred.Host, cred.Port)
				fmt.Printf("Username: %s\n", cred.Username)
				fmt.Printf("Auth Type: %s\n", cred.AuthType)
				fmt.Println("---------------------")
			}

			return nil
		},
	}

	return cmd
}
