package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zjpiazza/plantastic/internal/models"
)

// FormType represents the types of forms
type FormType int

const (
	FormTypeGarden FormType = iota
	FormTypeBed
	FormTypeTask
)

// FormModel represents a form for adding/editing items
type FormModel struct {
	formType     FormType
	title        string
	inputs       []textinput.Model
	focusIndex   int
	width        int
	height       int
	storage      Storage
	submitted    bool
	cancelled    bool
	errorMessage string

	// For editing existing items
	existingID       string
	existingParentID string // Garden ID for beds, Garden/Bed ID for tasks
}

// NewGardenForm creates a new form for adding/editing gardens
func NewGardenForm(storage Storage, width, height int) FormModel {
	m := FormModel{
		formType: FormTypeGarden,
		title:    "Add Garden",
		width:    width,
		height:   height,
		storage:  storage,
		inputs:   make([]textinput.Model, 3),
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

// NewBedForm creates a new form for adding/editing beds
func NewBedForm(storage Storage, gardenID string, width, height int) FormModel {
	m := FormModel{
		formType:         FormTypeBed,
		title:            "Add Bed",
		width:            width,
		height:           height,
		storage:          storage,
		inputs:           make([]textinput.Model, 5),
		existingParentID: gardenID,
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

// NewTaskForm creates a new form for adding/editing tasks
func NewTaskForm(storage Storage, gardenID string, bedID *string, width, height int) FormModel {
	m := FormModel{
		formType:         FormTypeTask,
		title:            "Add Task",
		width:            width,
		height:           height,
		storage:          storage,
		inputs:           make([]textinput.Model, 4),
		existingParentID: gardenID,
	}

	// Description input
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Description"
	m.inputs[0].Focus()
	m.inputs[0].Width = 40

	// Due date input
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Due Date (YYYY-MM-DD)"
	m.inputs[1].Width = 30
	m.inputs[1].SetValue(time.Now().Format("2006-01-02"))

	// Status input
	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "Status"
	m.inputs[2].Width = 30
	m.inputs[2].SetValue(models.TaskStatusPending)

	// Priority input
	m.inputs[3] = textinput.New()
	m.inputs[3].Placeholder = "Priority"
	m.inputs[3].Width = 30
	m.inputs[3].SetValue(models.PriorityMedium)

	return m
}

// SetGarden sets the form to edit an existing garden
func (m *FormModel) SetGarden(garden models.Garden) {
	m.title = "Edit Garden"
	m.existingID = garden.ID
	m.inputs[0].SetValue(garden.Name)
	m.inputs[1].SetValue(garden.Location)
	m.inputs[2].SetValue(garden.Description)
}

// SetBed sets the form to edit an existing bed
func (m *FormModel) SetBed(bed models.Bed) {
	m.title = "Edit Bed"
	m.existingID = bed.ID
	m.existingParentID = bed.GardenID
	m.inputs[0].SetValue(bed.Name)
	m.inputs[1].SetValue(bed.Type)
	m.inputs[2].SetValue(bed.Size)
	m.inputs[3].SetValue(bed.SoilType)
	m.inputs[4].SetValue(bed.Notes)
}

// SetTask sets the form to edit an existing task
func (m *FormModel) SetTask(task models.Task) {
	m.title = "Edit Task"
	m.existingID = task.ID
	m.existingParentID = task.GardenID
	m.inputs[0].SetValue(task.Description)
	m.inputs[1].SetValue(task.DueDate.Format("2006-01-02"))
	m.inputs[2].SetValue(task.Status)
	m.inputs[3].SetValue(task.Priority)
}

// Init initializes the form
func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles form events
func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

			switch m.formType {
			case FormTypeGarden:
				if err := m.submitGardenForm(); err != nil {
					m.errorMessage = err.Error()
					return m, nil
				}

			case FormTypeBed:
				if err := m.submitBedForm(); err != nil {
					m.errorMessage = err.Error()
					return m, nil
				}

			case FormTypeTask:
				if err := m.submitTaskForm(); err != nil {
					m.errorMessage = err.Error()
					return m, nil
				}
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
func (m *FormModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// submitGardenForm validates and submits the garden form
func (m *FormModel) submitGardenForm() error {
	name := strings.TrimSpace(m.inputs[0].Value())
	location := strings.TrimSpace(m.inputs[1].Value())
	description := strings.TrimSpace(m.inputs[2].Value())

	if name == "" {
		return fmt.Errorf("Garden name is required")
	}

	if m.existingID != "" {
		// Update existing garden
		garden, found := m.storage.GetGarden(m.existingID)
		if !found {
			return fmt.Errorf("Garden not found")
		}

		garden.Name = name
		garden.Location = location
		garden.Description = description
		garden.UpdatedAt = time.Now()

		return m.storage.UpdateGarden(garden)
	} else {
		// Create new garden
		garden := models.NewGarden(name, location, description)
		return m.storage.AddGarden(garden)
	}
}

// submitBedForm validates and submits the bed form
func (m *FormModel) submitBedForm() error {
	name := strings.TrimSpace(m.inputs[0].Value())
	bedType := strings.TrimSpace(m.inputs[1].Value())
	size := strings.TrimSpace(m.inputs[2].Value())
	soilType := strings.TrimSpace(m.inputs[3].Value())
	notes := strings.TrimSpace(m.inputs[4].Value())

	if name == "" {
		return fmt.Errorf("Bed name is required")
	}

	if m.existingParentID == "" {
		return fmt.Errorf("Garden ID is required")
	}

	if m.existingID != "" {
		// Update existing bed
		bed, found := m.storage.GetBed(m.existingID)
		if !found {
			return fmt.Errorf("Bed not found")
		}

		bed.Name = name
		bed.Type = bedType
		bed.Size = size
		bed.SoilType = soilType
		bed.Notes = notes
		bed.UpdatedAt = time.Now()

		return m.storage.UpdateBed(bed)
	} else {
		// Create new bed
		bed := models.NewBed(m.existingParentID, name, bedType, size, soilType, notes)
		return m.storage.AddBed(bed)
	}
}

// submitTaskForm validates and submits the task form
func (m *FormModel) submitTaskForm() error {
	description := strings.TrimSpace(m.inputs[0].Value())
	dueDateStr := strings.TrimSpace(m.inputs[1].Value())
	status := strings.TrimSpace(m.inputs[2].Value())
	priority := strings.TrimSpace(m.inputs[3].Value())

	if description == "" {
		return fmt.Errorf("Task description is required")
	}

	if m.existingParentID == "" {
		return fmt.Errorf("Garden ID is required")
	}

	// Parse due date
	dueDate, err := time.Parse("2006-01-02", dueDateStr)
	if err != nil {
		return fmt.Errorf("Invalid due date format, use YYYY-MM-DD")
	}

	// Get bed ID if applicable
	var bedID *string
	if len(m.existingParentID) > 0 {
		id := m.existingParentID
		bedID = &id
	}

	if m.existingID != "" {
		// Update existing task
		task, found := m.storage.GetTask(m.existingID)
		if !found {
			return fmt.Errorf("Task not found")
		}

		task.Description = description
		task.DueDate = dueDate
		task.Status = status
		task.Priority = priority
		task.UpdatedAt = time.Now()

		return m.storage.UpdateTask(task)
	} else {
		// Create new task
		task := models.NewTask(m.existingParentID, bedID, description, dueDate, status, priority)
		return m.storage.AddTask(task)
	}
}

// View renders the form
func (m FormModel) View() string {
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
func (m FormModel) Submitted() bool {
	return m.submitted
}

// Cancelled returns true if the form was cancelled
func (m FormModel) Cancelled() bool {
	return m.cancelled
}
