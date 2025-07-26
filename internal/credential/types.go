package credential

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type AuthType string

const (
	Password AuthType = "password"
	KeyFile  AuthType = "key"
)

type SSHCredential struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	AuthType  AuthType  `json:"auth_type"`
	Password  string    `json:"password,omitempty"`
	KeyPath   string    `json:"key_path,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GenerateID creates a unique ID for the credential
func GenerateID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
