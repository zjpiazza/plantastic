package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zjpiazza/plantastic/internal/models"
)

// GardenForm represents a form for adding/editing gardens
type GardenForm struct {
	title        string
	inputs       []textinput.Model
	focusIndex   int
	width        int
	height       int
	storage      Storage
	submitted    bool
	cancelled    bool
	errorMessage string
	onSave       func(models.Garden)

	// For editing existing gardens
	garden models.Garden
	isEdit bool
}

// Storage interface for garden operations
type Storage interface {
	GetGarden(id string) (models.Garden, bool)
	AddGarden(garden models.Garden) error
	UpdateGarden(garden models.Garden) error
}

// NewGardenForm creates a new form for adding/editing gardens
func NewGardenForm(storage Storage, width, height int, onSave func(models.Garden)) GardenForm {
	m := GardenForm{
		title:   "Add Garden",
		width:   width,
		height:  height,
		storage: storage,
		inputs:  make([]textinput.Model, 3),
		onSave:  onSave,
	}

	// Name input
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Name"
	m.inputs[0].Focus()
	m.inputs[0].Width = 30

	// Location input
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Location"
	m.inputs[1].Width = 30

	// Description input
	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "Description"
	m.inputs[2].Width = 40

	return m
}

// SetGarden sets the form to edit an existing garden
func (m *GardenForm) SetGarden(garden models.Garden) {
	m.title = "Edit Garden"
	m.isEdit = true
	m.garden = garden
	m.inputs[0].SetValue(garden.Name)
	m.inputs[1].SetValue(garden.Location)
	m.inputs[2].SetValue(garden.Description)
}

// Init initializes the form
func (m GardenForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles form events
func (m GardenForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, nil

		case "tab", "shift+tab", "up", "down":
			// Cycle focus through inputs
			s := msg.String()
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			} else if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			}

			for i := 0; i < len(m.inputs); i++ {
				if i == m.focusIndex {
					cmds = append(cmds, m.inputs[i].Focus())
				} else {
					m.inputs[i].Blur()
				}
			}

			return m, tea.Batch(cmds...)

		case "enter":
			// Submit the form
			m.errorMessage = ""

			if err := m.submitForm(); err != nil {
				m.errorMessage = err.Error()
				return m, nil
			}

			m.submitted = true
			return m, nil
		}
	}

	// Handle text input updates
	cmd := m.updateInputs(msg)
	return m, cmd
}

// updateInputs updates the text inputs
func (m *GardenForm) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// submitForm validates and submits the garden form
func (m *GardenForm) submitForm() error {
	name := strings.TrimSpace(m.inputs[0].Value())
	location := strings.TrimSpace(m.inputs[1].Value())
	description := strings.TrimSpace(m.inputs[2].Value())

	if name == "" {
		return fmt.Errorf("garden name is required")
	}

	var garden models.Garden

	if m.isEdit {
		// Update existing garden
		garden = m.garden
		garden.Name = name
		garden.Location = location
		garden.Description = description
		garden.UpdatedAt = time.Now()

		if err := m.storage.UpdateGarden(garden); err != nil {
			return fmt.Errorf("failed to update garden: %w", err)
		}
	} else {
		// Create new garden
		garden = models.NewGarden(name, location, description)
		if err := m.storage.AddGarden(garden); err != nil {
			return fmt.Errorf("failed to add garden: %w", err)
		}
	}

	// Call the onSave callback with the garden
	if m.onSave != nil {
		m.onSave(garden)
	}

	return nil
}

// View renders the form
func (m GardenForm) View() string {
	var b strings.Builder

	// Form title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#25A065")).
		Padding(1, 0, 1, 2)

	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Form inputs
	for i, input := range m.inputs {
		b.WriteString("  ")
		b.WriteString(input.View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n\n")
		}
	}

	// Error message
	if m.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF3B30")).
			Padding(1, 0)
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render("Error: " + m.errorMessage))
	}

	// Form controls
	controlsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Padding(2, 0)

	b.WriteString("\n\n")
	b.WriteString(controlsStyle.Render("TAB: Next field • SHIFT+TAB: Previous field • ENTER: Submit • ESC: Cancel"))

	return b.String()
}

// Submitted returns true if the form was submitted
func (m GardenForm) Submitted() bool {
	return m.submitted
}

// Cancelled returns true if the form was cancelled
func (m GardenForm) Cancelled() bool {
	return m.cancelled
}
