package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hemupadhyay26/ssh-cred-manager-cli/internal/credential"
	"github.com/spf13/cobra"
)

// connectToCredential dispatches to the correct connection method based on the OS.
// It uses tmux on Linux/macOS for a rich UI, and connects directly on Windows.
func connectToCredential(cred *credential.SSHCredential) error {
	if runtime.GOOS == "windows" {
		return connectDirectly(cred)
	}
	return connectWithTmux(cred)
}

// connectWithTmux uses tmux to create a session with a persistent status bar.
func connectWithTmux(cred *credential.SSHCredential) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("this feature requires `tmux` to be installed, but it was not found in your PATH.\n\nTo install it, use your system's package manager (e.g., 'sudo apt-get install tmux' or 'brew install tmux')")
	}

	// Create a unique session name to avoid conflicts.
	sessionName := fmt.Sprintf("ssh-cli-%s-%d", cred.Name, time.Now().Unix())

	// Construct the status bar text with some nice formatting.
	statusText := fmt.Sprintf(" #[bg=blue]#[fg=white] 🔗 %s (%s@%s) ", cred.Name, cred.Username, cred.Host)

	// Construct the full SSH command that tmux will execute as a single shell command string.
	var sshCmdBuilder strings.Builder
	if cred.AuthType == credential.Password {
		sshpassPath, err := exec.LookPath("sshpass")
		if err != nil {
			return fmt.Errorf("`sshpass` is not installed or not in your PATH. It is required for password-based authentication")
		}
		// Quote the password to handle special characters.
		sshCmdBuilder.WriteString(fmt.Sprintf("%s -p '%s' ", sshpassPath, cred.Password))
	}

	sshCmdBuilder.WriteString(fmt.Sprintf("ssh -p %d ", cred.Port))
	if cred.AuthType == credential.KeyFile {
		// Quote the key path.
		sshCmdBuilder.WriteString(fmt.Sprintf("-i '%s' ", cred.KeyPath))
	}
	// Quote the user@host argument.
	sshCmdBuilder.WriteString(fmt.Sprintf("'%s@%s' ", cred.Username, cred.Host))
	sshCmdBuilder.WriteString("-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null")

	// Build the full tmux command with chained commands.
	// This creates a new session, runs the ssh command, sets all the status bar options,
	// and finally attaches to the session, giving control to the user.
	tmuxArgs := []string{
		"new-session", "-s", sessionName, "-n", cred.Name, sshCmdBuilder.String(),
		"\\;", "set-option", "-t", sessionName, "status", "on",
		"\\;", "set-option", "-t", sessionName, "status-position", "top",
		"\\;", "set-option", "-t", sessionName, "status-right", statusText,
		"\\;", "set-option", "-t", sessionName, "status-style", "bg=default,fg=default",
		"\\;", "set-option", "-t", sessionName, "status-left", "''",
		"\\;", "attach-session", "-t", sessionName,
	}

	// The final command to run.
	cmd := exec.Command(tmuxPath, tmuxArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// connectDirectly connects to the SSH server without any UI wrappers.
// This is the fallback for Windows.
func connectDirectly(cred *credential.SSHCredential) error {
	fmt.Printf("Connecting to %s (%s@%s:%d)...\n", cred.Name, cred.Username, cred.Host, cred.Port)

	var sshCmd *exec.Cmd
	// Use 'nul' on Windows for the null device, '/dev/null' on others.
	nullDevice := "nul"
	if runtime.GOOS != "windows" {
		nullDevice = "/dev/null"
	}

	sshArgs := []string{
		"-p", strconv.Itoa(cred.Port),
		fmt.Sprintf("%s@%s", cred.Username, cred.Host),
		"-o", "StrictHostKeyChecking=no",
		"-o", fmt.Sprintf("UserKnownHostsFile=%s", nullDevice),
	}

	if cred.AuthType == credential.Password {
		// sshpass is not standard on Windows. For now, we can inform the user.
		return fmt.Errorf("password-based authentication is not currently supported for direct connections on Windows. Please use key-based authentication")
	}

	// Key-based authentication
	keyArgs := []string{"-i", cred.KeyPath}
	sshArgs = append(sshArgs, keyArgs...)
	sshCmd = exec.Command("ssh", sshArgs...)

	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	return sshCmd.Run()
}

// handleInteractiveConnect displays a menu for the user to select a credential to connect to.
func handleInteractiveConnect(store *credential.CredentialStore) error {
	creds := store.ListCredentials()
	if len(creds) == 0 {
		return fmt.Errorf("no credentials found. Use 'ssh-cli ssh save' to add one")
	}

	for {
		clearScreen()
		fmt.Println("Select a server to connect to:")
		fmt.Println("------------------------------")
		for i, cred := range creds {
			fmt.Printf("  %d. %s (%s@%s)\n", i+1, cred.Name, cred.Username, cred.Host)
		}
		fmt.Println("\nCommands:")
		fmt.Println("  1-N: Select a server to connect")
		fmt.Println("  q:   Quit")

		var choice string
		fmt.Print("\nEnter command: ")
		fmt.Scanln(&choice)

		if strings.ToLower(choice) == "q" {
			return nil
		}

		// Try to parse as number for selection
		if num, err := strconv.Atoi(choice); err == nil && num > 0 && num <= len(creds) {
			if err := connectToCredential(&creds[num-1]); err != nil {
				fmt.Printf("Connection failed: %v. Press Enter to continue.", err)
				fmt.Scanln()
			}
			continue // Continue the loop to show the menu again
		}

		fmt.Printf("Invalid selection '%s'. Press Enter to try again.", choice)
		fmt.Scanln()
	}
}

func NewConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect [name]",
		Short: "Connect to a saved SSH server",
		Long:  `Connects to an SSH server using the specified saved credential name. If no name is provided, an interactive menu will be shown.`,
		Args:  cobra.MaximumNArgs(1), // Allow 0 or 1 arg
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := credential.NewCredentialStore()
			if err != nil {
				return fmt.Errorf("failed to initialize credential store: %w", err)
			}

			// If a name is provided, connect directly
			if len(args) > 0 {
				name := args[0]
				cred, err := store.GetCredential(name)
				if err != nil {
					// GetCredential returns an error if not found, so this is sufficient.
					return err
				}
				return connectToCredential(cred)
			}

			return handleInteractiveConnect(store)
		},
	}
	return cmd
}
