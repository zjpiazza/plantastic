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

// BedForm represents a form for adding/editing beds
type BedForm struct {
	title        string
	inputs       []textinput.Model
	focusIndex   int
	width        int
	height       int
	storage      BedStorage
	submitted    bool
	cancelled    bool
	errorMessage string
	onSave       func(models.Bed)

	// For editing existing beds
	bed      models.Bed
	isEdit   bool
	gardenID string
}

// BedStorage interface for bed operations
type BedStorage interface {
	GetBed(id string) (models.Bed, bool)
	AddBed(bed models.Bed) error
	UpdateBed(bed models.Bed) error
}

// NewBedForm creates a new form for adding/editing beds
func NewBedForm(storage BedStorage, gardenID string, width, height int, onSave func(models.Bed)) BedForm {
	m := BedForm{
		title:    "Add Bed",
		width:    width,
		height:   height,
		storage:  storage,
		inputs:   make([]textinput.Model, 5),
		onSave:   onSave,
		gardenID: gardenID,
	}

	// Name input
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Name"
	m.inputs[0].Focus()
	m.inputs[0].Width = 30

	// Type input
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Type (e.g., Raised, In-Ground)"
	m.inputs[1].Width = 30

	// Size input
	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "Size"
	m.inputs[2].Width = 30

	// Soil type input
	m.inputs[3] = textinput.New()
	m.inputs[3].Placeholder = "Soil Type"
	m.inputs[3].Width = 30

	// Notes input
	m.inputs[4] = textinput.New()
	m.inputs[4].Placeholder = "Notes"
	m.inputs[4].Width = 40

	return m
}

// SetBed sets the form to edit an existing bed
func (m *BedForm) SetBed(bed models.Bed) {
	m.title = "Edit Bed"
	m.isEdit = true
	m.bed = bed
	m.gardenID = bed.GardenID
	m.inputs[0].SetValue(bed.Name)
	m.inputs[1].SetValue(bed.Type)
	m.inputs[2].SetValue(bed.Size)
	m.inputs[3].SetValue(bed.SoilType)
	m.inputs[4].SetValue(bed.Notes)
}

// Init initializes the form
func (m BedForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles form events
func (m BedForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *BedForm) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// submitForm validates and submits the bed form
func (m *BedForm) submitForm() error {
	name := strings.TrimSpace(m.inputs[0].Value())
	bedType := strings.TrimSpace(m.inputs[1].Value())
	size := strings.TrimSpace(m.inputs[2].Value())
	soilType := strings.TrimSpace(m.inputs[3].Value())
	notes := strings.TrimSpace(m.inputs[4].Value())

	if name == "" {
		return fmt.Errorf("bed name is required")
	}

	if m.gardenID == "" {
		return fmt.Errorf("garden ID is required")
	}

	var bed models.Bed

	if m.isEdit {
		// Update existing bed
		bed = m.bed
		bed.Name = name
		bed.Type = bedType
		bed.Size = size
		bed.SoilType = soilType
		bed.Notes = notes
		bed.UpdatedAt = time.Now()

		if err := m.storage.UpdateBed(bed); err != nil {
			return fmt.Errorf("failed to update bed: %w", err)
		}
	} else {
		// Create new bed
		bed = models.NewBed(m.gardenID, name, bedType, size, soilType, notes)
		if err := m.storage.AddBed(bed); err != nil {
			return fmt.Errorf("failed to add bed: %w", err)
		}
	}

	// Call the onSave callback with the bed
	if m.onSave != nil {
		m.onSave(bed)
	}

	return nil
}

// View renders the form
func (m BedForm) View() string {
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
func (m BedForm) Submitted() bool {
	return m.submitted
}

// Cancelled returns true if the form was cancelled
func (m BedForm) Cancelled() bool {
	return m.cancelled
}
