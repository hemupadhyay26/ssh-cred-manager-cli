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

func NewSaveWizardCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "wizard [user@host[:port] ...]",
		Short:   "Add one or more SSH credentials quickly",
		Aliases: []string{"w", "wiz"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("please provide at least one connection string (user@host[:port])")
			}

			store, err := credential.NewCredentialStore()
			if err != nil {
				return fmt.Errorf("failed to open credential store: %w", err)
			}

			reader := bufio.NewReader(os.Stdin)

			for _, connStr := range args {
				connStr = strings.TrimSpace(connStr)
				if connStr == "" {
					continue
				}

				// Parse connection string
				atIdx := strings.Index(connStr, "@")
				if atIdx == -1 {
					fmt.Printf("Invalid connection string: %s\n", connStr)
					continue
				}
				username := connStr[:atIdx]
				hostPort := connStr[atIdx+1:]
				port := 22
				host := hostPort
				colonIdx := strings.LastIndex(hostPort, ":")
				if colonIdx != -1 {
					host = hostPort[:colonIdx]
					portStr := hostPort[colonIdx+1:]
					if p, err := strconv.Atoi(portStr); err == nil {
						port = p
					}
				}

				// Prompt for connection name and ensure uniqueness
				var name string
				for {
					fmt.Printf("Enter name for %s (default: %s@%s): ", connStr, username, host)
					nameInput, _ := reader.ReadString('\n')
					nameInput = strings.TrimSpace(nameInput)
					if nameInput == "" {
						name = fmt.Sprintf("%s@%s", username, host)
					} else {
						name = nameInput
					}
					name = strings.ToLower(strings.TrimSpace(name))
					if name == "" {
						fmt.Println("Name cannot be empty. Please enter a valid name.")
						continue
					}
					// Check for uniqueness
					if existing, _ := store.GetCredential(name); existing != nil {
						fmt.Printf("A credential with the name '%s' already exists. Please enter a different name.\n", name)
						continue
					}
					break
				}

				// For fast mode, always use key auth and default key path
				authType := credential.KeyFile
				keyPath := os.ExpandEnv("$HOME/.ssh/id_rsa")

				id, err := credential.GenerateID()
				if err != nil {
					fmt.Printf("Failed to generate ID for %s: %v\n", connStr, err)
					continue
				}

				cred := credential.SSHCredential{
					ID:        id,
					Name:      name,
					Host:      host,
					Port:      port,
					Username:  username,
					AuthType:  authType,
					KeyPath:   keyPath,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				if err := store.SaveCredential(cred); err != nil {
					fmt.Printf("Failed to save %s: %v\n", connStr, err)
					continue
				}
				fmt.Printf("Saved: %s (%s@%s:%d)\n", name, username, host, port)
			}

			return nil
		},
	}
}
