package ssh

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hemupadhyay26/ssh-cred-manager-cli/internal/credential"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all saved SSH credentials",
		Aliases: []string{"ls", "l"},
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

			longOutput, _ := cmd.Flags().GetBool("long")

			if longOutput {
				fmt.Println("Saved SSH credentials (long output):")
				fmt.Println("---------------------")
				for i, cred := range credentials {
					fmt.Printf("[%d] Name: %s\n", i+1, cred.Name)
					fmt.Printf("    ID: %s\n", cred.ID)
					fmt.Printf("    Host: %s:%d\n", cred.Host, cred.Port)
					fmt.Printf("    Username: %s\n", cred.Username)
					fmt.Printf("    Auth Type: %s\n", cred.AuthType)
					fmt.Println("---------------------")
				}
				return nil
			}

			fmt.Println("Saved SSH credentials:")
			fmt.Println("---------------------")
			for i, cred := range credentials {
				fmt.Printf("[%d] Name: %s | ID: %s\n", i+1, cred.Name, cred.ID)
			}
			fmt.Println("---------------------")
			fmt.Print("Enter the numbers of the credentials you want to view (comma-separated, e.g. 1,3): ")
			var input string
			fmt.Scanln(&input)
			input = strings.TrimSpace(input)
			if input == "" {
				fmt.Println("No selection made.")
				return nil
			}
			selections := strings.Split(input, ",")
			selected := make(map[int]bool)
			for _, sel := range selections {
				sel = strings.TrimSpace(sel)
				if idx, err := strconv.Atoi(sel); err == nil && idx > 0 && idx <= len(credentials) {
					selected[idx-1] = true
				}
			}
			fmt.Println("\nSelected SSH credentials:")
			fmt.Println("---------------------")
			for i, cred := range credentials {
				if selected[i] {
					fmt.Printf("Name: %s\n", cred.Name)
					fmt.Printf("ID: %s\n", cred.ID)
					fmt.Printf("Host: %s:%d\n", cred.Host, cred.Port)
					fmt.Printf("Username: %s\n", cred.Username)
					fmt.Printf("Auth Type: %s\n", cred.AuthType)
					fmt.Println("---------------------")
				}
			}
			return nil
		},
	}

	// Add -l/--long flag for long output
	cmd.Flags().BoolP("long", "l", false, "Show detailed output (long format)")

	return cmd
}
