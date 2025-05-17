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

// TaskForm represents a form for adding/editing tasks
type TaskForm struct {
	title        string
	inputs       []textinput.Model
	focusIndex   int
	width        int
	height       int
	storage      TaskStorage
	submitted    bool
	cancelled    bool
	errorMessage string
	onSave       func(models.Task)

	// For editing existing tasks
	task     models.Task
	isEdit   bool
	gardenID string
	bedID    *string
	dueDate  time.Time
}

// TaskStorage interface for task operations
type TaskStorage interface {
	GetTask(id string) (models.Task, bool)
	AddTask(task models.Task) error
	UpdateTask(task models.Task) error
}

// NewTaskForm creates a new form for adding/editing tasks
func NewTaskForm(storage TaskStorage, gardenID string, bedID *string, width, height int, onSave func(models.Task)) TaskForm {
	m := TaskForm{
		title:    "Add Task",
		width:    width,
		height:   height,
		storage:  storage,
		inputs:   make([]textinput.Model, 3),
		onSave:   onSave,
		gardenID: gardenID,
		bedID:    bedID,
		dueDate:  time.Now().AddDate(0, 0, 7), // Default due date is one week from today
	}

	// Description input
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Description"
	m.inputs[0].Focus()
	m.inputs[0].Width = 40

	// Status input
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = fmt.Sprintf("Status (Pending, In Progress, Completed, etc.)")
	m.inputs[1].Width = 30
	m.inputs[1].SetValue(models.TaskStatusPending)

	// Priority input
	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = fmt.Sprintf("Priority (Low, Medium, High)")
	m.inputs[2].Width = 30
	m.inputs[2].SetValue(models.PriorityMedium)

	return m
}

// SetTask sets the form to edit an existing task
func (m *TaskForm) SetTask(task models.Task) {
	m.title = "Edit Task"
	m.isEdit = true
	m.task = task
	m.gardenID = task.GardenID
	m.bedID = task.BedID
	m.dueDate = task.DueDate
	m.inputs[0].SetValue(task.Description)
	m.inputs[1].SetValue(task.Status)
	m.inputs[2].SetValue(task.Priority)
}

// SetDueDate sets the due date for the task
func (m *TaskForm) SetDueDate(date time.Time) {
	m.dueDate = date
}

// Init initializes the form
func (m TaskForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles form events
func (m TaskForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *TaskForm) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// submitForm validates and submits the task form
func (m *TaskForm) submitForm() error {
	description := strings.TrimSpace(m.inputs[0].Value())
	status := strings.TrimSpace(m.inputs[1].Value())
	priority := strings.TrimSpace(m.inputs[2].Value())

	if description == "" {
		return fmt.Errorf("task description is required")
	}

	if m.gardenID == "" {
		return fmt.Errorf("garden ID is required")
	}

	// Validate status
	validStatus := false
	for _, s := range []string{models.TaskStatusPending, models.TaskStatusInProgress, models.TaskStatusCompleted, models.TaskStatusOverdue, models.TaskStatusCancelled} {
		if status == s {
			validStatus = true
			break
		}
	}
	if !validStatus {
		return fmt.Errorf("invalid status: use Pending, In Progress, Completed, Overdue, or Cancelled")
	}

	// Validate priority
	validPriority := false
	for _, p := range []string{models.PriorityLow, models.PriorityMedium, models.PriorityHigh} {
		if priority == p {
			validPriority = true
			break
		}
	}
	if !validPriority {
		return fmt.Errorf("invalid priority: use Low, Medium, or High")
	}

	var task models.Task

	if m.isEdit {
		// Update existing task
		task = m.task
		task.Description = description
		task.Status = status
		task.Priority = priority
		task.DueDate = m.dueDate
		task.UpdatedAt = time.Now()

		if err := m.storage.UpdateTask(task); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}
	} else {
		// Create new task
		task = models.NewTask(m.gardenID, m.bedID, description, m.dueDate, status, priority)
		if err := m.storage.AddTask(task); err != nil {
			return fmt.Errorf("failed to add task: %w", err)
		}
	}

	// Call the onSave callback with the task
	if m.onSave != nil {
		m.onSave(task)
	}

	return nil
}

// View renders the form
func (m TaskForm) View() string {
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

	// Due date display
	dueDateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#007AFF")).
		Padding(1, 0)

	b.WriteString("\n\n")
	b.WriteString(dueDateStyle.Render(fmt.Sprintf("  Due Date: %s", m.dueDate.Format("2006-01-02"))))
	b.WriteString("\n  (Press D to change due date)")

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
func (m TaskForm) Submitted() bool {
	return m.submitted
}

// Cancelled returns true if the form was cancelled
func (m TaskForm) Cancelled() bool {
	return m.cancelled
}
