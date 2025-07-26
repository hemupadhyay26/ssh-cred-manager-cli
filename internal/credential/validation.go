package credential

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

func (c *SSHCredential) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("name cannot be empty")
	}

	if strings.TrimSpace(c.Host) == "" {
		return errors.New("host cannot be empty")
	}

	if _, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port)); err != nil {
		return fmt.Errorf("invalid host or port: %v", err)
	}

	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	if strings.TrimSpace(c.Username) == "" {
		return errors.New("username cannot be empty")
	}

	switch c.AuthType {
	case Password:
		if strings.TrimSpace(c.Password) == "" {
			return errors.New("password cannot be empty when using password authentication")
		}
	case KeyFile:
		if strings.TrimSpace(c.KeyPath) == "" {
			return errors.New("key path cannot be empty when using key authentication")
		}
		if _, err := os.Stat(c.KeyPath); os.IsNotExist(err) {
			return errors.New("SSH key file does not exist")
		}
	default:
		return errors.New("invalid authentication type")
	}

	return nil
}
