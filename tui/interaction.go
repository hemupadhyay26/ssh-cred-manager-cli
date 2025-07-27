package tui

import (
	"os"
	"path/filepath"

	"github.com/hemupadhyay26/ssh-cred-manager-cli/internal/credential"
)

// store is the global CredentialStore instance.
var store *credential.CredentialStore

// init initializes the CredentialStore.
func init() {
	var err error
	store, err = credential.NewCredentialStore()
	if err != nil {
		panic(err) // In a real app, handle this error gracefully
	}
}

// toTuiCredential converts a credential.SSHCredential to tui.SSHCredential.
func toTuiCredential(cred credential.SSHCredential) SSHCredential {
	return SSHCredential{
		ID:        cred.ID,
		Name:      cred.Name,
		Host:      cred.Host,
		Port:      cred.Port,
		Username:  cred.Username,
		AuthType:  string(cred.AuthType),
		Password:  cred.Password,
		KeyPath:   cred.KeyPath,
		CreatedAt: cred.CreatedAt,
		UpdatedAt: cred.UpdatedAt,
	}
}

// toCredentialSSHCredential converts a tui.SSHCredential to credential.SSHCredential.
func toCredentialSSHCredential(cred SSHCredential) credential.SSHCredential {
	return credential.SSHCredential{
		ID:        cred.ID,
		Name:      cred.Name,
		Host:      cred.Host,
		Port:      cred.Port,
		Username:  cred.Username,
		AuthType:  credential.AuthType(cred.AuthType),
		Password:  cred.Password,
		KeyPath:   cred.KeyPath,
		CreatedAt: cred.CreatedAt,
		UpdatedAt: cred.UpdatedAt,
	}
}

// GetCredentials retrieves all SSH credentials.
func GetCredentials() ([]SSHCredential, error) {
	creds := store.ListCredentials()
	tuiCreds := make([]SSHCredential, len(creds))
	for i, cred := range creds {
		tuiCreds[i] = toTuiCredential(cred)
	}
	return tuiCreds, nil
}

// SaveCredential adds or updates a credential.
func SaveCredential(cred SSHCredential) error {
	return store.SaveCredential(toCredentialSSHCredential(cred))
}

// DeleteCredential deletes a credential by name.
func DeleteCredential(name string) error {
	return store.DeleteCredential(name)
}

// UpdateCredential updates a credential by name.
func UpdateCredential(name string, cred SSHCredential) error {
	return store.UpdateCredential(name, toCredentialSSHCredential(cred))
}

// RenameCredential renames a credential.
func RenameCredential(oldName, newName string) error {
	return store.RenameCredential(oldName, newName)
}

// GetDefaultKeyPath returns the path to the default SSH key if it exists.
func GetDefaultKeyPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Check for common key types in order of preference
	keyTypes := []string{"id_rsa", "id_ed25519", "id_ecdsa", "id_dsa"}
	for _, keyType := range keyTypes {
		keyPath := filepath.Join(homeDir, ".ssh", keyType)
		if _, err := os.Stat(keyPath); err == nil {
			return keyPath, nil
		}
	}

	// Return default even if it doesn't exist
	return filepath.Join(homeDir, ".ssh", "id_rsa"), nil
}
