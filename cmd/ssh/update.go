package ssh

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hemupadhyay26/ssh-cred-manager-cli/internal/credential"
	"github.com/spf13/cobra"
)

func NewUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "update [credential name or id]",
		Short:   "Update an existing SSH credential",
		Aliases: []string{"u", "up"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("Updating SSH credential...\n")
			reader := bufio.NewReader(os.Stdin)
			store, err := credential.NewCredentialStore()
			if err != nil {
				return fmt.Errorf("failed to open credential store: %w", err)
			}

			var cred *credential.SSHCredential
			var nameOrID string

			if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
				nameOrID = strings.TrimSpace(args[0])
				cred, err = store.GetCredential(nameOrID)
				if err != nil {
					return fmt.Errorf("credential not found: %w", err)
				}
			} else {
				// List all credentials
				creds := store.ListCredentials()
				if len(creds) == 0 {
					return fmt.Errorf("no credentials found")
				}
				fmt.Println("Available credentials:")
				for i, c := range creds {
					fmt.Printf("[%d] %s (ID: %s)\n", i+1, c.Name, c.ID)
				}
				fmt.Print("Select credential by number: ")
				choiceStr, _ := reader.ReadString('\n')
				choiceStr = strings.TrimSpace(choiceStr)
				choice, err := strconv.Atoi(choiceStr)
				if err != nil || choice < 1 || choice > len(creds) {
					return fmt.Errorf("invalid selection")
				}
				cred = &creds[choice-1]
				nameOrID = cred.Name
			}

			// Print old credential
			fmt.Println("\nCurrent credential values:")
			fmt.Printf("Host: %s\n", cred.Host)
			fmt.Printf("Port: %d\n", cred.Port)
			fmt.Printf("Username: %s\n", cred.Username)
			fmt.Printf("AuthType: %s\n", cred.AuthType)
			if cred.AuthType == credential.Password {
				fmt.Printf("Password: (hidden)\n")
			} else {
				fmt.Printf("KeyPath: %s\n", cred.KeyPath)
			}
			fmt.Println("Leave blank to keep current value.")

			fmt.Printf("New Host [%s]: ", cred.Host)
			host, _ := reader.ReadString('\n')
			host = strings.TrimSpace(host)
			if host != "" {
				cred.Host = host
			}

			fmt.Printf("New Port [%d]: ", cred.Port)
			portStr, _ := reader.ReadString('\n')
			portStr = strings.TrimSpace(portStr)
			if portStr != "" {
				if port, err := strconv.Atoi(portStr); err == nil {
					cred.Port = port
				}
			}

			fmt.Printf("New Username [%s]: ", cred.Username)
			username, _ := reader.ReadString('\n')
			username = strings.TrimSpace(username)
			if username != "" {
				cred.Username = username
			}

			fmt.Printf("New AuthType [%s] [password/key]: ", cred.AuthType)
			authType, _ := reader.ReadString('\n')
			authType = strings.TrimSpace(authType)
			if authType != "" {
				cred.AuthType = credential.AuthType(authType)
			}

			if cred.AuthType == credential.Password {
				fmt.Printf("New Password (leave blank to keep): ")
				password, _ := reader.ReadString('\n')
				password = strings.TrimSpace(password)
				if password != "" {
					cred.Password = password
				}
			} else {
				fmt.Printf("New KeyPath [%s]: ", cred.KeyPath)
				keyPath, _ := reader.ReadString('\n')
				keyPath = strings.TrimSpace(keyPath)
				if keyPath != "" {
					cred.KeyPath = keyPath
				}
			}

			cred.UpdatedAt = time.Now()

			if err := store.UpdateCredential(nameOrID, *cred); err != nil {
				return fmt.Errorf("failed to update credential: %w", err)
			}

			fmt.Println("Credential updated successfully.")
			return nil
		},
	}
}
