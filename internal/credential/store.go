package credential

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CredentialStore manages SSH credentials.
type CredentialStore struct {
	Credentials []SSHCredential `json:"credentials"`
	filepath    string
}

// NewCredentialStore initializes and loads the store.
func NewCredentialStore() (*CredentialStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	storePath := filepath.Join(homeDir, ".ssh-cred-manager", "credentials.json")
	if err := os.MkdirAll(filepath.Dir(storePath), 0700); err != nil {
		return nil, err
	}

	store := &CredentialStore{
		filepath: storePath,
	}

	if err := store.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return store, nil
}

// load reads credentials from the JSON file.
func (s *CredentialStore) load() error {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.Credentials = []SSHCredential{}
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &s.Credentials)
}

// save writes current credentials to disk.
func (s *CredentialStore) save() error {
	data, err := json.MarshalIndent(s.Credentials, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filepath, data, 0600)
}

// SaveCredential creates or replaces a credential.
func (s *CredentialStore) SaveCredential(cred SSHCredential) error {
	if err := cred.Validate(); err != nil {
		return err
	}

	for i, existing := range s.Credentials {
		if existing.Name == cred.Name {
			s.Credentials[i] = cred
			return s.save()
		}
	}

	s.Credentials = append(s.Credentials, cred)
	return s.save()
}

// ListCredentials returns all credentials.
func (s *CredentialStore) ListCredentials() []SSHCredential {
	return s.Credentials
}

// GetCredential finds a credential by exact name.
func (s *CredentialStore) GetCredential(name string) (*SSHCredential, error) {
	for _, cred := range s.Credentials {
		if cred.Name == name {
			return &cred, nil
		}
	}
	return nil, fmt.Errorf("credential not found: %s", name)
}

// DeleteCredential deletes a credential by name.
func (s *CredentialStore) DeleteCredential(name string) error {
	for i, cred := range s.Credentials {
		if cred.Name == name {
			s.Credentials = append(s.Credentials[:i], s.Credentials[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("credential not found: %s", name)
}

// UpdateCredential updates a credential by name.
func (s *CredentialStore) UpdateCredential(name string, cred SSHCredential) error {
	if err := cred.Validate(); err != nil {
		return err
	}

	for i, existing := range s.Credentials {
		if existing.Name == name {
			s.Credentials[i] = cred
			return s.save()
		}
	}
	return fmt.Errorf("credential not found: %s", name)
}

// FindCredentialsByName performs case-insensitive partial search.
func (s *CredentialStore) FindCredentialsByName(name string) []SSHCredential {
	var matches []SSHCredential
	name = strings.ToLower(name)
	for _, cred := range s.Credentials {
		if strings.Contains(strings.ToLower(cred.Name), name) {
			matches = append(matches, cred)
		}
	}
	return matches
}

// Exists checks if a credential with the given name exists, optionally excluding a specific name.
func (s *CredentialStore) Exists(name string, exclude ...string) bool {
	for _, cred := range s.Credentials {
		if cred.Name == name {
			// Check if this is the excluded name
			for _, ex := range exclude {
				if ex == name {
					return false
				}
			}
			return true
		}
	}
	return false
}

// RenameCredential renames an existing credential after checking for uniqueness.
func (s *CredentialStore) RenameCredential(oldName, newName string) error {
	if newName == "" {
		return fmt.Errorf("new name cannot be empty")
	}
	// Check if newName is unique, excluding oldName
	if s.Exists(newName, oldName) {
		return fmt.Errorf("credential with name '%s' already exists", newName)
	}
	for i, cred := range s.Credentials {
		if cred.Name == oldName {
			s.Credentials[i].Name = newName
			return s.save()
		}
	}
	return fmt.Errorf("credential not found: %s", oldName)
}

// Count returns the total number of credentials.
func (s *CredentialStore) Count() int {
	return len(s.Credentials)
}

// ClearAllCredentials wipes all credentials (use carefully).
func (s *CredentialStore) ClearAllCredentials() error {
	s.Credentials = []SSHCredential{}
	return s.save()
}
