package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SSHCredential is a local copy of the credential structure for TUI use.
type SSHCredential struct {
	ID        string
	Name      string
	Host      string
	Port      int
	Username  string
	AuthType  string
	Password  string
	KeyPath   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Model represents the state of the TUI.
type Model struct {
	credentials    []SSHCredential // List of SSH credentials
	cursor         int             // Selected credential index in list view
	viewDetails    bool            // Whether to show details of selected credential
	adding         bool            // Whether in add credential mode
	editing        bool            // Whether in edit credential mode
	renaming       bool            // Whether in rename credential mode
	deleting       bool            // Whether in delete confirmation mode
	input          string          // Input buffer for current field
	inputField     string          // Current field being edited
	inputIndex     int             // Index of the input field
	inputs         []string        // Inputs for all fields during add/edit
	errorMsg       string          // Error message to display
	deleteConfirm  bool            // Tracks y/n input for delete confirmation
	defaultKeyPath string          // Default SSH key path
}

// msg types for updating the TUI state.
type msg interface{}

type credentialsMsg []SSHCredential
type errorMsg string
type defaultKeyMsg string

// inputFields defines the fields for adding/editing a credential.
var inputFields = []string{"name", "host", "port", "username", "auth_type", "password/key_path"}

// Define lipgloss styles for dark theme
var (
	// General style with dark background
	baseStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1C2526")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(1, 2)

	// Header style for title
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00A7A7")).
			Bold(true).
			MarginBottom(1)

	// Error message style
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			MarginBottom(1)

	// Selected item in list
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#2E2E2E")).
				PaddingLeft(4)

	// Unselected item in list
	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B0B0B0")).
			PaddingLeft(2)

	// Active input field
	activeInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				PaddingLeft(2)

	// Inactive input field
	inactiveInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#B0B0B0")).
				PaddingLeft(2)

	// Label in details view
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4"))

	// Value in details view
	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	// Prompt style for rename/delete
	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFC107")).
			MarginBottom(1)

	// Instruction text
	instructionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A0A0A0")).
				MarginTop(1)
)

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchCredentials, getDefaultKeyPath)
}

// Update handles incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.deleting {
			return m.handleDeleteConfirm(msg)
		}
		if m.adding || m.editing {
			return m.handleInput(msg)
		}
		if m.renaming {
			return m.handleRenameInput(msg)
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down":
			if m.cursor < len(m.credentials)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			if !m.viewDetails && len(m.credentials) > 0 {
				m.viewDetails = true
			} else {
				m.viewDetails = false
			}
			return m, nil
		case "a":
			m.adding = true
			m.inputs = []string{"", "", "22", "", "key", ""} // Default port: 22, auth_type: key
			m.inputIndex = 0
			m.inputField = inputFields[0]
			m.input = m.inputs[0]
			m.errorMsg = ""
			return m, nil
		case "e":
			if len(m.credentials) > 0 && m.cursor < len(m.credentials) {
				m.editing = true
				m.inputs = []string{
					m.credentials[m.cursor].Name,
					m.credentials[m.cursor].Host,
					strconv.Itoa(m.credentials[m.cursor].Port),
					m.credentials[m.cursor].Username,
					m.credentials[m.cursor].AuthType,
					ifThen(m.credentials[m.cursor].AuthType == "password", m.credentials[m.cursor].Password, m.credentials[m.cursor].KeyPath),
				}
				m.inputIndex = 0
				m.inputField = inputFields[0]
				m.input = m.inputs[0]
				m.errorMsg = ""
				return m, nil
			}
		case "r":
			if len(m.credentials) > 0 && m.cursor < len(m.credentials) {
				m.renaming = true
				m.input = ""
				m.inputIndex = 0
				m.inputField = "new_name"
				m.errorMsg = ""
				return m, nil
			}
		case "d":
			if len(m.credentials) > 0 && m.cursor < len(m.credentials) {
				m.deleting = true
				m.deleteConfirm = false
				m.errorMsg = ""
				return m, nil
			}
		}
	case credentialsMsg:
		m.credentials = msg
		if m.cursor >= len(m.credentials) {
			m.cursor = max(0, len(m.credentials)-1)
		}
		return m, nil
	case errorMsg:
		m.errorMsg = string(msg)
		return m, nil
	case defaultKeyMsg:
		m.defaultKeyPath = string(msg)
		return m, nil
	}
	return m, nil
}

// handleInput processes input for adding or editing credentials.
func (m Model) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.inputIndex > 0 {
			m.inputs[m.inputIndex] = m.input
			m.inputIndex--
			m.inputField = inputFields[m.inputIndex]
			m.input = m.inputs[m.inputIndex]
		}
		return m, nil
	case "down":
		// Skip password/key_path for add mode if auth_type is key
		maxIndex := len(inputFields) - 1
		if m.adding && m.inputs[4] == "key" {
			maxIndex = len(inputFields) - 2
		}
		if m.inputIndex < maxIndex {
			m.inputs[m.inputIndex] = m.input
			m.inputIndex++
			m.inputField = inputFields[m.inputIndex]
			m.input = m.inputs[m.inputIndex]
		}
		return m, nil
	case "enter":
		m.inputs[m.inputIndex] = m.input
		cred, err := m.createCredential()
		if err != nil {
			m.errorMsg = err.Error()
			return m, nil // Keep form open
		}
		cmd := addCredential(cred)
		if m.editing {
			cmd = updateCredential(m.credentials[m.cursor].Name, cred)
		}
		m.adding = false
		m.editing = false
		m.input = ""
		m.inputIndex = 0
		m.inputField = inputFields[0]
		m.inputs = nil
		m.errorMsg = ""
		return m, cmd
	case "ctrl+c", "esc":
		m.adding = false
		m.editing = false
		m.input = ""
		m.inputIndex = 0
		m.inputField = inputFields[0]
		m.inputs = nil
		m.errorMsg = ""
		return m, nil
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.input += string(msg.Runes)
		}
	}
	return m, nil
}

// handleRenameInput processes input for renaming a credential.
func (m Model) handleRenameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if m.input == "" {
			m.errorMsg = "new name cannot be empty"
			return m, nil // Keep rename mode open
		}
		newName := m.input // Capture the input
		// Attempt to rename and check for errors
		err := RenameCredential(m.credentials[m.cursor].Name, newName)
		if err != nil {
			m.errorMsg = fmt.Sprintf("failed to rename credential: %v", err)
			return m, nil // Keep rename mode open
		}
		// Success: exit rename mode and refresh credentials
		m.renaming = false
		m.input = ""
		m.inputIndex = 0
		m.inputField = inputFields[0]
		m.errorMsg = ""
		return m, func() tea.Msg { return fetchCredentials() }
	case tea.KeyCtrlC, tea.KeyEsc:
		m.renaming = false
		m.input = ""
		m.inputIndex = 0
		m.inputField = inputFields[0]
		m.errorMsg = ""
		return m, nil
	case tea.KeyBackspace:
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	case tea.KeyRunes:
		m.input += string(msg.Runes)
	}
	return m, nil
}

// handleDeleteConfirm processes y/n input for delete confirmation.
func (m Model) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.deleting = false
		m.deleteConfirm = false
		return m, deleteCredential(m.credentials[m.cursor].Name)
	case "n", "N", "ctrl+c", "esc":
		m.deleting = false
		m.deleteConfirm = false
		m.errorMsg = ""
		return m, nil
	default:
		return m, nil
	}
}

// createCredential constructs an SSHCredential from inputs.
func (m Model) createCredential() (SSHCredential, error) {
	if len(m.inputs) != len(inputFields) {
		return SSHCredential{}, fmt.Errorf("incomplete input")
	}
	port, err := strconv.Atoi(m.inputs[2])
	if err != nil || port <= 0 {
		return SSHCredential{}, fmt.Errorf("invalid port: %s", m.inputs[2])
	}
	authType := m.inputs[4]
	if authType != "password" && authType != "key" {
		return SSHCredential{}, fmt.Errorf("auth type must be 'password' or 'key'")
	}
	if m.inputs[0] == "" {
		return SSHCredential{}, fmt.Errorf("name cannot be empty")
	}
	if m.inputs[1] == "" {
		return SSHCredential{}, fmt.Errorf("host cannot be empty")
	}
	if m.inputs[3] == "" {
		return SSHCredential{}, fmt.Errorf("username cannot be empty")
	}
	if authType == "key" && m.inputs[5] == "" && m.editing {
		return SSHCredential{}, fmt.Errorf("key path cannot be empty")
	}
	keyPath := m.inputs[5]
	if authType == "key" && m.adding && m.inputs[5] == "" {
		if m.defaultKeyPath == "" {
			return SSHCredential{}, fmt.Errorf("no default SSH key found and key path not provided")
		}
		keyPath = m.defaultKeyPath
	}
	return SSHCredential{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:      m.inputs[0],
		Host:      m.inputs[1],
		Port:      port,
		Username:  m.inputs[3],
		AuthType:  authType,
		Password:  ifThen(authType == "password", m.inputs[5], ""),
		KeyPath:   ifThen(authType == "key", keyPath, ""),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// View renders the interface based on the current state.
func (m Model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Render("SSH Credential Manager") + "\n\n")

	// Error message
	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg)) + "\n\n")
	}

	// Delete confirmation
	if m.deleting {
		b.WriteString(promptStyle.Render(fmt.Sprintf("Delete credential '%s'? [y/n]", m.credentials[m.cursor].Name)) + "\n")
		return baseStyle.Render(b.String())
	}

	// Add/Edit form
	if m.adding || m.editing {
		b.WriteString(headerStyle.Render(fmt.Sprintf("%s Credential", map[bool]string{true: "Edit", false: "Add"}[m.editing])) + "\n")
		for i, field := range inputFields {
			if m.adding && field == "password/key_path" && m.inputs[4] == "key" {
				continue // Skip password/key_path for key auth in add mode
			}
			value := m.inputs[i]
			prefix := "  "
			style := inactiveInputStyle
			if i == m.inputIndex {
				prefix = "> "
				value = m.input
				style = activeInputStyle
			}
			b.WriteString(style.Render(fmt.Sprintf("%s%s: %s", prefix, field, value)) + "\n")
		}
		b.WriteString(instructionStyle.Render("[Up/Down] to navigate fields, [Enter] to submit, [Esc] to cancel") + "\n")
		return baseStyle.Render(b.String())
	}

	// Rename prompt
	if m.renaming {
		b.WriteString(promptStyle.Render(fmt.Sprintf("Renaming credential '%s' - new name: %s", m.credentials[m.cursor].Name, m.input)) + "\n")
		b.WriteString(instructionStyle.Render("[Enter] to submit, [Esc] to cancel") + "\n")
		return baseStyle.Render(b.String())
	}

	// Credential list or details
	if len(m.credentials) == 0 {
		b.WriteString(itemStyle.Render("No credentials found.") + "\n")
	} else if m.viewDetails && m.cursor < len(m.credentials) {
		cred := m.credentials[m.cursor]
		b.WriteString(labelStyle.Render("ID: ") + valueStyle.Render(cred.ID) + "\n")
		b.WriteString(labelStyle.Render("Name: ") + valueStyle.Render(cred.Name) + "\n")
		b.WriteString(labelStyle.Render("Host: ") + valueStyle.Render(cred.Host) + "\n")
		b.WriteString(labelStyle.Render("Port: ") + valueStyle.Render(strconv.Itoa(cred.Port)) + "\n")
		b.WriteString(labelStyle.Render("Username: ") + valueStyle.Render(cred.Username) + "\n")
		b.WriteString(labelStyle.Render("Auth Type: ") + valueStyle.Render(cred.AuthType) + "\n")
		if cred.AuthType == "password" {
			b.WriteString(labelStyle.Render("Password: ") + valueStyle.Render(cred.Password) + "\n")
		} else if cred.AuthType == "key" {
			b.WriteString(labelStyle.Render("Key Path: ") + valueStyle.Render(cred.KeyPath) + "\n")
		}
		b.WriteString(labelStyle.Render("Created At: ") + valueStyle.Render(cred.CreatedAt.Format(time.RFC3339)) + "\n")
		b.WriteString(labelStyle.Render("Updated At: ") + valueStyle.Render(cred.UpdatedAt.Format(time.RFC3339)) + "\n")
	} else {
		for i, cred := range m.credentials {
			cursor := " "
			style := itemStyle
			if i == m.cursor {
				cursor = ">"
				style = selectedItemStyle
			}
			b.WriteString(style.Render(fmt.Sprintf("%s %s (%s:%d)", cursor, cred.Name, cred.Host, cred.Port)) + "\n")
		}
	}

	// Instructions
	b.WriteString(instructionStyle.Render("[Up/Down] to navigate, [Enter] to toggle details, [a] to add, [e] to edit, [r] to rename, [d] to delete, [q] to quit") + "\n")

	return baseStyle.Render(b.String())
}

// fetchCredentials retrieves credentials from interaction.go.
func fetchCredentials() tea.Msg {
	creds, err := GetCredentials()
	if err != nil {
		return errorMsg(fmt.Sprintf("failed to fetch credentials: %v", err))
	}
	return credentialsMsg(creds)
}

// getDefaultKeyPath retrieves the default SSH key path.
func getDefaultKeyPath() tea.Msg {
	path, err := GetDefaultKeyPath()
	if err != nil {
		return defaultKeyMsg("")
	}
	return defaultKeyMsg(path)
}

// addCredential saves a new credential.
func addCredential(cred SSHCredential) tea.Cmd {
	return func() tea.Msg {
		if err := SaveCredential(cred); err != nil {
			return errorMsg(fmt.Sprintf("failed to add credential: %v", err))
		}
		return fetchCredentials()
	}
}

// updateCredential updates an existing credential.
func updateCredential(name string, cred SSHCredential) tea.Cmd {
	return func() tea.Msg {
		if err := UpdateCredential(name, cred); err != nil {
			return errorMsg(fmt.Sprintf("failed to update credential: %v", err))
		}
		return fetchCredentials()
	}
}

// deleteCredential deletes a credential by name.
func deleteCredential(name string) tea.Cmd {
	return func() tea.Msg {
		if err := DeleteCredential(name); err != nil {
			return errorMsg(fmt.Sprintf("failed to delete credential: %v", err))
		}
		return fetchCredentials()
	}
}

// renameCredential renames a credential.
func renameCredential(oldName, newName string) tea.Cmd {
	return func() tea.Msg {
		if err := RenameCredential(oldName, newName); err != nil {
			return errorMsg(fmt.Sprintf("failed to rename credential: %v", err))
		}
		return fetchCredentials()
	}
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ifThen returns a if condition is true, else b.
func ifThen(condition bool, a, b string) string {
	if condition {
		return a
	}
	return b
}

// Run starts the Bubble Tea program.
func Run() error {
	p := tea.NewProgram(Model{})
	if err := p.Start(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}
	return nil
}
