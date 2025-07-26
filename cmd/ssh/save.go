package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hemupadhyay26/ssh-cred-manager-cli/internal/credential"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func promptForNewName(store *credential.CredentialStore, originalName string) string {
	for {
		newName := promptForInput(fmt.Sprintf("Connection name '%s' already exists. Enter new name", originalName))
		if newName == "" {
			fmt.Println("Name cannot be empty. Try again.")
			continue
		}

		// Check if the new name also exists
		if cred, _ := store.GetCredential(newName); cred == nil {
			return newName
		}
		fmt.Printf("Connection name '%s' also exists. Try a different name.\n", newName)
	}
}

// promptForInput reads user input from stdin
func promptForInput(prompt string) string {
	fmt.Printf("%s: ", prompt)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

// promptForPassword reads password input securely without echoing
func promptForPassword(prompt string) string {
	fmt.Printf("%s: ", prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add newline after password input
	if err != nil {
		return ""
	}
	return string(password)
}

// getDefaultKeyPath returns the default SSH key path
func getDefaultKeyPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Check for common key types in order of preference
	keyTypes := []string{"id_rsa", "id_ed25519", "id_ecdsa", "id_dsa"}
	for _, keyType := range keyTypes {
		keyPath := filepath.Join(homeDir, ".ssh", keyType)
		if _, err := os.Stat(keyPath); err == nil {
			return keyPath
		}
	}

	// Return default even if it doesn't exist
	return filepath.Join(homeDir, ".ssh", "id_rsa")
}

func NewSaveCmd() *cobra.Command {
	var (
		name     string
		host     string
		port     int = 22 // Default port
		username string
		password string
		keyPath  string
		authType string = string(credential.KeyFile) // Default auth type
	)

	cmd := &cobra.Command{
		Use:     "save",
		Short:   "Save new SSH credentials",
		Aliases: []string{"s", "add", "a"},
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := credential.NewCredentialStore()
			if err != nil {
				return fmt.Errorf("failed to initialize credential store: %w", err)
			}

			println("Add new SSH credential")
			// Interactive prompts for missing required fields
			if name == "" {
				for {
					name = promptForInput("Enter connection name")
					name = strings.ToLower(strings.TrimSpace(name)) // Normalize here

					if name == "" {
						fmt.Println("Name cannot be empty. Try again.")
						continue
					}
					if cred, _ := store.GetCredential(name); cred == nil {
						break
					}
					name = promptForNewName(store, name)
					name = strings.ToLower(strings.TrimSpace(name)) // Normalize again after prompt
					break
				}
			} else {
				// Check if name from flag already exists
				name = strings.ToLower(strings.TrimSpace(name)) // Normalize here
				if cred, _ := store.GetCredential(name); cred != nil {
					name = promptForNewName(store, name)
					name = strings.ToLower(strings.TrimSpace(name)) // Normalize again after prompt
				}
			}

			if host == "" {
				host = promptForInput("Enter host address")
			}

			if username == "" {
				username = promptForInput("Enter username")
			}

			// Only prompt for auth type if explicitly set to empty
			if authType == "" {
				fmt.Println("Authentication type (password/key)")
				authType = promptForInput("Enter auth type")
			}

			var auth credential.AuthType
			switch authType {
			case "password":
				auth = credential.Password
				if password == "" {
					password = promptForPassword("Enter password")
				}
			case "key":
				auth = credential.KeyFile
				if keyPath == "" {
					defaultKey := getDefaultKeyPath()
					keyPath = defaultKey
					fmt.Printf("Using default key: %s\n", defaultKey)
				}
				// Verify key exists
				if _, err := os.Stat(keyPath); os.IsNotExist(err) {
					return fmt.Errorf("SSH key file not found: %s", keyPath)
				}
			default:
				return fmt.Errorf("invalid authentication type: use 'password' or 'key'")
			}

			// Validate inputs
			if name == "" || host == "" || username == "" {
				return fmt.Errorf("required fields cannot be empty")
			}

			id, err := credential.GenerateID()
			if err != nil {
				return fmt.Errorf("failed to generate unique ID: %w", err)
			}

			now := time.Now()
			cred := credential.SSHCredential{
				ID:        id,
				Name:      name, // Using potentially renamed connection
				Host:      host,
				Port:      port,
				Username:  username,
				AuthType:  auth,
				Password:  password,
				KeyPath:   keyPath,
				CreatedAt: now,
				UpdatedAt: now,
			}

			if err := store.SaveCredential(cred); err != nil {
				return fmt.Errorf("failed to save credential: %w", err)
			}

			fmt.Printf("Successfully saved SSH credential for %s\n", name)
			return nil
		},
	}

	// Add flags with defaults
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the SSH connection (required)")
	cmd.Flags().StringVarP(&host, "host", "H", "", "Host address (required)")
	cmd.Flags().IntVarP(&port, "port", "p", 22, "SSH port")
	cmd.Flags().StringVarP(&username, "user", "u", "", "SSH username (required)")
	cmd.Flags().StringVarP(&password, "password", "P", "", "SSH password (for password auth)")
	cmd.Flags().StringVarP(&keyPath, "key", "k", getDefaultKeyPath(), "SSH private key path")
	cmd.Flags().StringVarP(&authType, "auth-type", "a", "key", "Authentication type (password/key)")

	return cmd
}
