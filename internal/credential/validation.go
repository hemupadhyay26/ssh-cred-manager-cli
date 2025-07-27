package credential

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

func (c *SSHCredential) Validate() error {
	// Trim all key fields to avoid whitespace issues
	c.Name = strings.TrimSpace(c.Name)
	c.Host = strings.TrimSpace(c.Host)
	c.Username = strings.TrimSpace(c.Username)
	c.Password = strings.TrimSpace(c.Password)
	c.KeyPath = strings.TrimSpace(c.KeyPath)

	if c.Name == "" {
		return errors.New("name cannot be empty")
	}

	if c.Host == "" {
		return errors.New("host cannot be empty")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	if _, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port)); err != nil {
		return fmt.Errorf("invalid host or port: %v", err)
	}

	if c.Username == "" {
		return errors.New("username cannot be empty")
	}

	switch c.AuthType {
	case Password:
		if c.Password == "" {
			return errors.New("password cannot be empty when using password authentication")
		}
	case KeyFile:
		if c.KeyPath == "" {
			return errors.New("key path cannot be empty when using key authentication")
		}
		if statErr := fileExists(c.KeyPath); statErr != nil {
			return statErr
		}
	default:
		return errors.New("invalid authentication type")
	}

	return nil
}

// fileExists checks if the file exists and is accessible
func fileExists(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("SSH key file does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("unable to access SSH key file: %v", err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected a file but got directory: %s", path)
	}
	return nil
}
