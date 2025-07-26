package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/hemupadhyay26/ssh-cred-manager-cli/internal/credential"
	"github.com/spf13/cobra"
)

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func showCredentialDetails(cred *credential.SSHCredential, showSensitive bool) {
	fmt.Printf("\nConnection Details:\n")
	fmt.Printf("------------------\n")
	fmt.Printf("Name: %s\n", cred.Name)
	fmt.Printf("Host: %s:%d\n", cred.Host, cred.Port)
	fmt.Printf("Username: %s\n", cred.Username)
	fmt.Printf("Auth Type: %s\n", cred.AuthType)
	if showSensitive && cred.AuthType == credential.Password {
		fmt.Printf("Password: %s\n", cred.Password)
	}
	if cred.AuthType == credential.KeyFile {
		fmt.Printf("Key Path: %s\n", cred.KeyPath)
	}
}

func confirmDelete(cred *credential.SSHCredential) bool {
	showCredentialDetails(cred, false)
	fmt.Print("\nDo you want to delete this credential? (y/n): ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

func handleMultipleDelete(store *credential.CredentialStore, creds []credential.SSHCredential) error {
	selectedIndices := make(map[int]bool)
	showSensitive := false
	var deletedCreds []credential.SSHCredential

	for {
		clearScreen()
		fmt.Println("\nAvailable credentials:")
		fmt.Println("--------------------")
		for i, cred := range creds {
			checked := "[ ]"
			if selectedIndices[i] {
				checked = "[x]"
			}
			fmt.Printf("  %s %d. %s (%s@%s)\n", checked, i+1, cred.Name, cred.Username, cred.Host)
		}

		// Display the command menu
		fmt.Println("\nCommands:")
		fmt.Println("  1-N: Toggle selection")
		fmt.Println("  v:   View credential details")
		fmt.Println("  t:   Toggle sensitive information")
		fmt.Println("  d:   Delete selected credentials")
		fmt.Println("  q:   Quit without deleting")

		var choice string
		fmt.Print("\nEnter command: ")
		fmt.Scanln(&choice)

		switch strings.ToLower(choice) {
		case "q":
			return nil
		case "v":
			clearScreen()
			fmt.Print("Enter credential number to view: ")
			var num int
			fmt.Scanln(&num)
			if num > 0 && num <= len(creds) {
				showCredentialDetails(&creds[num-1], showSensitive)
			}
			fmt.Print("\nPress Enter to continue...")
			fmt.Scanln()
		case "t":
			showSensitive = !showSensitive
			fmt.Printf("Sensitive information is now %s\n", map[bool]string{true: "shown", false: "hidden"}[showSensitive])
		case "d":
			if len(selectedIndices) == 0 {
				fmt.Println("No credentials selected")
				continue
			}
			clearScreen()
			fmt.Println("Selected credentials to delete:")
			fmt.Println("------------------------------")
			for i := range selectedIndices {
				showCredentialDetails(&creds[i], false)
			}
			fmt.Print("\nAre you sure you want to delete selected credentials? (y/n): ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(confirm) == "y" {
				clearScreen()
				fmt.Println("Deleting credentials...")
				fmt.Println("---------------------")
				for i := range selectedIndices {
					if err := store.DeleteCredential(creds[i].Name); err != nil {
						return fmt.Errorf("failed to delete credential %s: %w", creds[i].Name, err)
					}
					deletedCreds = append(deletedCreds, creds[i])
				}
				fmt.Println("\nSuccessfully deleted credentials:")
				fmt.Println("--------------------------------")
				for _, cred := range deletedCreds {
					showCredentialDetails(&cred, false)
				}
				return nil
			}
		default:
			// Try to parse as number for selection
			var num int
			if _, err := fmt.Sscanf(choice, "%d", &num); err == nil && num > 0 && num <= len(creds) {
				selectedIndices[num-1] = !selectedIndices[num-1]
			}
		}
	}
}

func NewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [name]",
		Short:   "Delete saved SSH credential(s)",
		Aliases: []string{"del", "rm", "d"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := credential.NewCredentialStore()
			if err != nil {
				return fmt.Errorf("failed to initialize credential store: %w", err)
			}

			// If name provided, delete single credential
			if len(args) > 0 {
				name := args[0]
				cred, err := store.GetCredential(name)
				if err != nil || cred == nil {
					return fmt.Errorf("credential not found: %s", name)
				}

				if !confirmDelete(cred) {
					fmt.Println("Deletion cancelled")
					return nil
				}

				clearScreen()
				fmt.Println("Deleting credential:")
				fmt.Println("-------------------")
				showCredentialDetails(cred, false)

				if err := store.DeleteCredential(name); err != nil {
					return fmt.Errorf("failed to delete credential: %w", err)
				}

				fmt.Println("\nSuccessfully deleted credential:")
				fmt.Println("------------------------------")
				showCredentialDetails(cred, false)
				return nil
			}

			// No name provided - show interactive selection
			creds := store.ListCredentials()
			if len(creds) == 0 {
				return fmt.Errorf("no credentials found")
			}

			return handleMultipleDelete(store, creds)
		},
	}

	return cmd
}
