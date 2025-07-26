package credential

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CredentialStore struct {
	Credentials []SSHCredential `json:"credentials"`
	filepath    string
}

func (s *CredentialStore) FindCredentialsByName(name string) []SSHCredential {
	var matches []SSHCredential
	for _, cred := range s.Credentials {
		if strings.Contains(strings.ToLower(cred.Name), strings.ToLower(name)) {
			matches = append(matches, cred)
		}
	}
	return matches
}

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

	if _, err := os.Stat(storePath); !os.IsNotExist(err) {
		if err := store.load(); err != nil {
			return nil, err
		}
	}

	return store, nil
}

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

func (s *CredentialStore) load() error {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s)
}

func (s *CredentialStore) save() error {
	data, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filepath, data, 0600)
}

// ListCredentials returns all stored credentials
func (s *CredentialStore) ListCredentials() []SSHCredential {
	return s.Credentials
}

// GetCredential returns a credential by name
func (s *CredentialStore) GetCredential(name string) (*SSHCredential, error) {
	for _, cred := range s.Credentials {
		if cred.Name == name {
			return &cred, nil
		}
	}
	return nil, fmt.Errorf("credential not found: %s", name)
}

// DeleteCredential removes a credential by name
func (s *CredentialStore) DeleteCredential(name string) error {
	for i, cred := range s.Credentials {
		if cred.Name == name {
			// Remove the credential from the slice
			s.Credentials = append(s.Credentials[:i], s.Credentials[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("credential not found: %s", name)
}

// UpdateCredential updates an existing credential
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
