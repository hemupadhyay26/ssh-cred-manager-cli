package ssh

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/hemupadhyay26/ssh-cred-manager-cli/internal/credential"
	"github.com/spf13/cobra"
)

// NewConnectCmd returns a cobra command for connecting via SSH.
func NewConnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "connect",
		Short:   "Connect to an SSH server using a saved credential",
		Aliases: []string{"c", "conn"},
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter credential name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)

			store, err := credential.NewCredentialStore()
			if err != nil {
				return fmt.Errorf("failed to open credential store: %w", err)
			}

			cred, err := store.GetCredential(name)
			if err != nil {
				return fmt.Errorf("credential not found: %w", err)
			}

			fmt.Printf("Connecting to %s@%s:%d...\n", cred.Username, cred.Host, cred.Port)

			sshArgs := []string{
				"-p", strconv.Itoa(cred.Port),
				"-i", cred.KeyPath,
				fmt.Sprintf("%s@%s", cred.Username, cred.Host),
				"-o", "StrictHostKeyChecking=no",
				"-o", "UserKnownHostsFile=/dev/null",
			}

			cmdExec := exec.Command("ssh", sshArgs...)
			cmdExec.Stdin = os.Stdin
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr

			return cmdExec.Run()
		},
	}
}
