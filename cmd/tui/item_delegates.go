package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zjpiazza/plantastic/internal/models"
)

// GardenItem represents a garden in the list
type GardenItem struct {
	garden models.Garden
}

func (i GardenItem) Title() string       { return i.garden.Name }
func (i GardenItem) Description() string { return i.garden.Description }
func (i GardenItem) FilterValue() string { return i.garden.Name }

// BedItem represents a bed in the list
type BedItem struct {
	bed models.Bed
}

func (i BedItem) Title() string       { return i.bed.Name }
func (i BedItem) Description() string { return fmt.Sprintf("%s - %s", i.bed.Type, i.bed.Notes) }
func (i BedItem) FilterValue() string { return i.bed.Name }

// GardenDelegate is a custom delegate for rendering gardens in the list
type GardenDelegate struct {
	styles list.DefaultItemStyles
}

// NewGardenDelegate creates a new garden delegate with custom styles
func NewGardenDelegate(styles list.DefaultItemStyles) list.ItemDelegate {
	styles.NormalTitle = styles.NormalTitle.Copy().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	styles.NormalDesc = styles.NormalDesc.Copy().
		Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#999999"})

	styles.SelectedTitle = styles.SelectedTitle.Copy().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#25A065")).
		Foreground(lipgloss.Color("#25A065")).
		Bold(true)

	styles.SelectedDesc = styles.SelectedDesc.Copy().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#25A065")).
		Foreground(lipgloss.Color("#25A065"))

	return &GardenDelegate{styles: styles}
}

func (d GardenDelegate) Height() int { return 2 }

func (d GardenDelegate) Spacing() int { return 1 }

func (d GardenDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d GardenDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	gardenItem, ok := item.(GardenItem)
	if !ok {
		return
	}

	garden := gardenItem.garden
	if index == m.Index() {
		fmt.Fprint(w, d.styles.SelectedTitle.Render(garden.Name))
		fmt.Fprintf(w, "\n%s", d.styles.SelectedDesc.Render(garden.Description))
	} else {
		fmt.Fprint(w, d.styles.NormalTitle.Render(garden.Name))
		fmt.Fprintf(w, "\n%s", d.styles.NormalDesc.Render(garden.Description))
	}
}

// BedDelegate is a custom delegate for rendering beds in the list
type BedDelegate struct {
	styles list.DefaultItemStyles
}

// NewBedDelegate creates a new bed delegate with custom styles
func NewBedDelegate(styles list.DefaultItemStyles) list.ItemDelegate {
	styles.NormalTitle = styles.NormalTitle.Copy().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	styles.NormalDesc = styles.NormalDesc.Copy().
		Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#999999"})

	styles.SelectedTitle = styles.SelectedTitle.Copy().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#25A065")).
		Foreground(lipgloss.Color("#25A065")).
		Bold(true)

	styles.SelectedDesc = styles.SelectedDesc.Copy().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#25A065")).
		Foreground(lipgloss.Color("#25A065"))

	return &BedDelegate{styles: styles}
}

func (d BedDelegate) Height() int { return 2 }

func (d BedDelegate) Spacing() int { return 1 }

func (d BedDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d BedDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	bedItem, ok := item.(BedItem)
	if !ok {
		return
	}

	bed := bedItem.bed
	desc := fmt.Sprintf("%s - %s", bed.Type, bed.Notes)
	if index == m.Index() {
		fmt.Fprint(w, d.styles.SelectedTitle.Render(bed.Name))
		fmt.Fprintf(w, "\n%s", d.styles.SelectedDesc.Render(desc))
	} else {
		fmt.Fprint(w, d.styles.NormalTitle.Render(bed.Name))
		fmt.Fprintf(w, "\n%s", d.styles.NormalDesc.Render(desc))
	}
}

// TaskItem implements list.Item for Task
type TaskItem struct {
	task models.Task
}

func (t TaskItem) Title() string { return t.task.Description }
func (t TaskItem) Description() string {
	return fmt.Sprintf("Due: %s | %s", t.task.DueDate.Format("2006-01-02"), t.task.Status)
}
func (t TaskItem) FilterValue() string { return t.task.Description }

// NewTaskDelegate creates a custom delegate for task items
func NewTaskDelegate(styles list.DefaultItemStyles) list.ItemDelegate {
	return TaskDelegate{
		styles: styles,
	}
}

// TaskDelegate is a custom delegate for task items
type TaskDelegate struct {
	styles list.DefaultItemStyles
}

func (d TaskDelegate) Height() int { return 2 }

func (d TaskDelegate) Spacing() int { return 1 }

func (d TaskDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d TaskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	taskItem, ok := item.(TaskItem)
	if !ok {
		return
	}

	task := taskItem.task
	textColor := lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}

	// Style task based on status
	statusStyle := lipgloss.NewStyle()
	switch task.Status {
	case models.TaskStatusPending:
		statusStyle = statusStyle.Foreground(lipgloss.Color("#FF9500")) // Orange
	case models.TaskStatusInProgress:
		statusStyle = statusStyle.Foreground(lipgloss.Color("#007AFF")) // Blue
	case models.TaskStatusCompleted:
		statusStyle = statusStyle.Foreground(lipgloss.Color("#4CD964")) // Green
	case models.TaskStatusOverdue:
		statusStyle = statusStyle.Foreground(lipgloss.Color("#FF3B30")) // Red
	}

	statusText := statusStyle.Render(task.Status)

	if index == m.Index() {
		fmt.Fprint(w, d.styles.SelectedTitle.Render(task.Description))
		fmt.Fprintf(w, "\n%s", d.styles.SelectedDesc.Render(fmt.Sprintf("Due: %s | %s",
			task.DueDate.Format("2006-01-02"), statusText)))
	} else {
		fmt.Fprint(w, d.styles.NormalTitle.Render(task.Description))
		fmt.Fprintf(w, "\n%s", lipgloss.NewStyle().Foreground(textColor).Render(fmt.Sprintf("Due: %s | %s",
			task.DueDate.Format("2006-01-02"), statusText)))
	}
}
