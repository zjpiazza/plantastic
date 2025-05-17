package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"

	"github.com/zjpiazza/plantastic/internal/models"
)

// Define custom tabs component since the bubbles/tabs package might not be available
type tabsModel struct {
	Tabs        []string
	ActiveTab   int
	ActiveColor string
}

func newTabs(tabs []string) tabsModel {
	return tabsModel{
		Tabs:        tabs,
		ActiveTab:   0,
		ActiveColor: "205",
	}
}

func (t *tabsModel) FocusLeft() {
	if t.ActiveTab > 0 {
		t.ActiveTab--
	} else {
		t.ActiveTab = len(t.Tabs) - 1
	}
}

func (t *tabsModel) FocusRight() {
	if t.ActiveTab < len(t.Tabs)-1 {
		t.ActiveTab++
	} else {
		t.ActiveTab = 0
	}
}

func (t tabsModel) View() string {
	var tabViews []string

	tabStyle := lipgloss.NewStyle().
		Padding(0, 3).
		MarginRight(1).
		Border(lipgloss.ThickBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("240"))

	activeTabStyle := tabStyle.
		Foreground(lipgloss.Color(t.ActiveColor)).
		Border(lipgloss.ThickBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(t.ActiveColor)).
		Bold(true)

	for i, tab := range t.Tabs {
		if i == t.ActiveTab {
			tabViews = append(tabViews, activeTabStyle.Render(tab))
		} else {
			tabViews = append(tabViews, tabStyle.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)
}

// PlantasticClient is the API client for Plantastic.
type PlantasticClient struct {
	// Replace with actual API configuration.
	BaseURL string
	APIKey  string
}

// GetGardens fetches gardens from the API.
func (c *PlantasticClient) GetGardens() ([]models.Garden, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/gardens", c.BaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("API connection error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error: %s", response.Status)
	}

	var gardens []models.Garden
	if err := json.NewDecoder(response.Body).Decode(&gardens); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return gardens, nil
}

// GetBeds fetches beds from the API.
func (c *PlantasticClient) GetBeds(gardenID string) ([]models.Bed, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/beds", c.BaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("API connection error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error: %s", response.Status)
	}

	var beds []models.Bed
	if err := json.NewDecoder(response.Body).Decode(&beds); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return beds, nil
}

// GetTasks fetches tasks from the API.
func (c *PlantasticClient) GetTasks(gardenID, bedID string) ([]models.Task, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/tasks", c.BaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("API connection error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error: %s", response.Status)
	}

	var tasks []models.Task
	if err := json.NewDecoder(response.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return tasks, nil
}

// CompleteTask marks a task as completed.
func (c *PlantasticClient) CompleteTask(taskID string) error {
	// Placeholder - replace with actual API call.
	fmt.Println("Completing task:", taskID)
	return nil
}

// Model represents the application state.
type Model struct {
	client    PlantasticClient
	tabs      tabsModel
	activeTab int

	// Data
	gardens []models.Garden
	beds    []models.Bed
	tasks   []models.Task

	// Selected items
	selectedGarden *models.Garden
	selectedBed    *models.Bed
	selectedTask   *models.Task

	// Tables
	gardenTable table.Model
	bedTable    table.Model
	taskTable   table.Model

	// UI State
	spinner spinner.Model
	loading bool
	err     error
	help    help.Model
	keys    keyMap
	width   int
	height  int
}

// Tab indices
const (
	GardenTab = 0
	BedTab    = 1
	TaskTab   = 2
)

// keyMap defines key mappings.
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Tab      key.Binding
	Select   key.Binding
	Complete key.Binding
	Refresh  key.Binding
	Quit     key.Binding
	Help     key.Binding

	// Stores the active tab to show contextual help
	activeTab int
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "previous tab"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next tab"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Complete: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "complete task"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),

	// Default active tab
	activeTab: GardenTab,
}

// FullHelp implements help.KeyMap.
func (k keyMap) FullHelp() [][]key.Binding {
	navigation := []key.Binding{k.Up, k.Down, k.Left, k.Right}
	actions := []key.Binding{k.Select, k.Refresh}

	// Only add the Complete binding when on the Tasks tab
	if k.activeTab == TaskTab {
		actions = append(actions, k.Complete)
	}

	system := []key.Binding{k.Help, k.Quit}

	return [][]key.Binding{navigation, actions, system}
}

// ShortHelp implements help.KeyMap.
func (k keyMap) ShortHelp() []key.Binding {
	result := []key.Binding{k.Tab, k.Select, k.Refresh, k.Quit}

	// Only add the Complete binding when on the Tasks tab
	if k.activeTab == TaskTab {
		result = append(result, k.Complete)
	}

	return result
}

// NewModel initializes the model.
func NewModel() Model {
	// Gardens table
	gardenCols := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: 20},
		{Title: "Location", Width: 20},
		{Title: "Description", Width: 30},
	}

	// Beds table
	bedCols := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: 20},
		{Title: "Type", Width: 15},
		{Title: "Size", Width: 10},
		{Title: "Soil Type", Width: 15},
	}

	// Tasks table
	taskCols := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Description", Width: 30},
		{Title: "Due Date", Width: 12},
		{Title: "Status", Width: 10},
		{Title: "Priority", Width: 10},
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Improved table styling with rounded borders and better colors
	t := table.DefaultStyles()
	t.Header = t.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true). // Only show bottom border for header
		Bold(false)
	t.Selected = t.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("205")).
		Bold(false)

	// Create tables with consistent styling
	gardenTable := table.New(
		table.WithColumns(gardenCols),
		table.WithFocused(true),
		table.WithHeight(10),
		table.WithStyles(t),
	)

	bedTable := table.New(
		table.WithColumns(bedCols),
		table.WithFocused(true),
		table.WithHeight(10),
		table.WithStyles(t),
	)

	taskTable := table.New(
		table.WithColumns(taskCols),
		table.WithFocused(true),
		table.WithHeight(10),
		table.WithStyles(t),
	)

	tabsModel := newTabs([]string{"Gardens", "Beds", "Tasks"})

	// Create a copy of the keys with the default active tab
	localKeys := keys
	localKeys.activeTab = GardenTab

	return Model{
		client: PlantasticClient{
			BaseURL: "http://localhost:8080",
			APIKey:  "your-api-key",
		},
		tabs:        tabsModel,
		activeTab:   GardenTab,
		gardenTable: gardenTable,
		bedTable:    bedTable,
		taskTable:   taskTable,
		spinner:     s,
		loading:     true,
		help:        help.New(),
		keys:        localKeys,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchGardens(),
		m.fetchBeds(""),      // Get all beds initially
		m.fetchTasks("", ""), // Get all tasks initially
	)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			switch m.activeTab {
			case GardenTab:
				return m, m.fetchGardens()
			case BedTab:
				gardenID := ""
				if m.selectedGarden != nil {
					gardenID = m.selectedGarden.ID
				}
				return m, m.fetchBeds(gardenID)
			case TaskTab:
				gardenID := ""
				bedID := ""
				if m.selectedGarden != nil {
					gardenID = m.selectedGarden.ID
				}
				if m.selectedBed != nil {
					bedID = m.selectedBed.ID
				}
				return m, m.fetchTasks(gardenID, bedID)
			}

		case key.Matches(msg, m.keys.Tab), key.Matches(msg, m.keys.Right):
			m.activeTab = (m.activeTab + 1) % 3
			tabs := m.tabs
			tabs.FocusRight()
			m.tabs = tabs

			// Update the active tab in keyMap
			keys := m.keys
			keys.activeTab = m.activeTab
			m.keys = keys

		case key.Matches(msg, m.keys.Left):
			m.activeTab = (m.activeTab + 2) % 3
			tabs := m.tabs
			tabs.FocusLeft()
			m.tabs = tabs

			// Update the active tab in keyMap
			keys := m.keys
			keys.activeTab = m.activeTab
			m.keys = keys

		case key.Matches(msg, m.keys.Complete):
			if m.activeTab == TaskTab && m.selectedTask != nil {
				return m, m.completeTask(m.selectedTask.ID)
			}
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		verticalMargins := 15 // Space for header, tabs, help and details
		tableHeight := m.height - verticalMargins

		// Ensure minimum viable height
		if tableHeight < 5 {
			tableHeight = 5
		}

		// Update table heights
		m.gardenTable.SetHeight(tableHeight)
		m.bedTable.SetHeight(tableHeight)
		m.taskTable.SetHeight(tableHeight)

		m.help.Width = msg.Width

		// Calculate appropriate table widths based on columns
		// We don't set the width of the tables directly as that can cause alignment issues
		// Instead, we make sure columns don't exceed available width
		availableWidth := msg.Width - 4 // Account for borders and margins

		// Adjust garden table columns
		gardenColsWidth := 0
		for _, col := range m.gardenTable.Columns() {
			gardenColsWidth += col.Width
		}
		if gardenColsWidth > availableWidth {
			ratio := float64(availableWidth) / float64(gardenColsWidth)
			cols := m.gardenTable.Columns()
			for i := range cols {
				cols[i].Width = int(float64(cols[i].Width) * ratio)
			}
			m.gardenTable.SetColumns(cols)
		}

		// Adjust bed table columns
		bedColsWidth := 0
		for _, col := range m.bedTable.Columns() {
			bedColsWidth += col.Width
		}
		if bedColsWidth > availableWidth {
			ratio := float64(availableWidth) / float64(bedColsWidth)
			cols := m.bedTable.Columns()
			for i := range cols {
				cols[i].Width = int(float64(cols[i].Width) * ratio)
			}
			m.bedTable.SetColumns(cols)
		}

		// Adjust task table columns
		taskColsWidth := 0
		for _, col := range m.taskTable.Columns() {
			taskColsWidth += col.Width
		}
		if taskColsWidth > availableWidth {
			ratio := float64(availableWidth) / float64(taskColsWidth)
			cols := m.taskTable.Columns()
			for i := range cols {
				cols[i].Width = int(float64(cols[i].Width) * ratio)
			}
			m.taskTable.SetColumns(cols)
		}

	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case fetchGardensMsg:
		m.loading = false
		m.gardens = msg
		m.updateGardenTable()

	case fetchBedsMsg:
		m.loading = false
		m.beds = msg
		m.updateBedTable()

	case fetchTasksMsg:
		m.loading = false
		m.tasks = msg
		m.updateTaskTable()

	case fetchErrMsg:
		m.loading = false
		m.err = msg

	case completeTaskMsg:
		gardenID := ""
		bedID := ""
		if m.selectedGarden != nil {
			gardenID = m.selectedGarden.ID
		}
		if m.selectedBed != nil {
			bedID = m.selectedBed.ID
		}
		return m, m.fetchTasks(gardenID, bedID)
	}

	// Handle tab-specific updates
	switch m.activeTab {
	case GardenTab:
		m.gardenTable, cmd = m.gardenTable.Update(msg)
		cmds = append(cmds, cmd)

		// Update selected garden
		if len(m.gardens) > 0 && m.gardenTable.Cursor() < len(m.gardens) {
			selected := m.gardens[m.gardenTable.Cursor()]
			m.selectedGarden = &selected

			// When garden is selected, refresh beds for this garden
			if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, m.keys.Select) {
				cmds = append(cmds, m.fetchBeds(selected.ID))

				// Also update tasks if we have a garden and bed selected
				if m.selectedBed != nil && m.selectedBed.GardenID == selected.ID {
					cmds = append(cmds, m.fetchTasks(selected.ID, m.selectedBed.ID))
				} else {
					cmds = append(cmds, m.fetchTasks(selected.ID, ""))
				}
			}
		}

	case BedTab:
		m.bedTable, cmd = m.bedTable.Update(msg)
		cmds = append(cmds, cmd)

		// Update selected bed
		if len(m.beds) > 0 && m.bedTable.Cursor() < len(m.beds) {
			selected := m.beds[m.bedTable.Cursor()]
			m.selectedBed = &selected

			// When bed is selected, refresh tasks for this bed
			if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, m.keys.Select) {
				gardenID := ""
				if m.selectedGarden != nil {
					gardenID = m.selectedGarden.ID
				}
				cmds = append(cmds, m.fetchTasks(gardenID, selected.ID))
			}
		}

	case TaskTab:
		m.taskTable, cmd = m.taskTable.Update(msg)
		cmds = append(cmds, cmd)

		// Update selected task
		if len(m.tasks) > 0 && m.taskTable.Cursor() < len(m.tasks) {
			selected := m.tasks[m.taskTable.Cursor()]
			m.selectedTask = &selected
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m Model) View() string {
	// Center the content in the terminal
	doc := strings.Builder{}
	width := m.width

	// Title with a nicer style
	title := "🌱 Plantastic"
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Margin(1, 0).
		Padding(0, 2).
		Width(width).
		Align(lipgloss.Center)
	doc.WriteString(titleStyle.Render(title) + "\n")

	// Loading view
	if m.loading {
		loadingText := fmt.Sprintf("%s Loading data...", m.spinner.View())
		loadingStyle := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("240")).
			Italic(true)
		doc.WriteString(loadingStyle.Render(loadingText) + "\n\n")
		return doc.String()
	}

	// Error view
	if m.err != nil {
		errorText := fmt.Sprintf("Error: %v", m.err)
		errorStyle := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("9")).
			Bold(true)
		doc.WriteString(errorStyle.Render(errorText) + "\n\n")
		helpStyle := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center)
		doc.WriteString(helpStyle.Render(m.help.View(m.keys)))
		return doc.String()
	}

	// Tabs with improved styling
	tabsView := m.tabs.View()
	tabsStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		MarginBottom(1)
	doc.WriteString(tabsStyle.Render(tabsView) + "\n")

	// Get the current table
	var tableView string
	switch m.activeTab {
	case GardenTab:
		tableView = m.gardenTable.View()
	case BedTab:
		tableView = m.bedTable.View()
	case TaskTab:
		tableView = m.taskTable.View()
	}

	// Center the table with proper styling
	tableStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center)

	// The table is already rendered by the table component, we're just centering it here
	doc.WriteString(tableStyle.Render(tableView) + "\n")

	// Details section
	detailsView := m.renderDetails()
	if detailsView != "" {
		detailsStyle := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("246")).
			MarginTop(1).
			MarginBottom(1)
		doc.WriteString(detailsStyle.Render(detailsView) + "\n")
	}

	// Help
	helpStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("241")).
		MarginTop(1)
	doc.WriteString(helpStyle.Render(m.help.View(m.keys)))

	return doc.String()
}

// renderDetails renders details for the selected item
func (m Model) renderDetails() string {
	var details string

	switch m.activeTab {
	case GardenTab:
		if m.selectedGarden != nil {
			details = fmt.Sprintf(
				"Garden: %s\nLocation: %s\nDescription: %s",
				lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("205")).
					Render(m.selectedGarden.Name),
				m.selectedGarden.Location,
				m.selectedGarden.Description,
			)
		}
	case BedTab:
		if m.selectedBed != nil {
			details = fmt.Sprintf(
				"Bed: %s\nType: %s | Size: %s | Soil: %s\nNotes: %s",
				lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("205")).
					Render(m.selectedBed.Name),
				m.selectedBed.Type,
				m.selectedBed.Size,
				m.selectedBed.SoilType,
				m.selectedBed.Notes,
			)
		}
	case TaskTab:
		if m.selectedTask != nil {
			// Add a colored status indicator
			statusStyle := lipgloss.NewStyle()
			switch strings.ToLower(m.selectedTask.Status) {
			case "pending":
				statusStyle = statusStyle.Foreground(lipgloss.Color("11")) // Yellow
			case "completed":
				statusStyle = statusStyle.Foreground(lipgloss.Color("10")) // Green
			case "overdue":
				statusStyle = statusStyle.Foreground(lipgloss.Color("9")) // Red
			default:
				statusStyle = statusStyle.Foreground(lipgloss.Color("12")) // Blue
			}

			details = fmt.Sprintf(
				"Task: %s\nStatus: %s | Priority: %s | Due: %s",
				lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("205")).
					Render(m.selectedTask.Description),
				statusStyle.Render(m.selectedTask.Status),
				m.selectedTask.Priority,
				m.selectedTask.DueDate,
			)

			// Add a hint to complete the task if it's not already completed
			if strings.ToLower(m.selectedTask.Status) != "completed" {
				details += "\n" + lipgloss.NewStyle().
					Italic(true).
					Foreground(lipgloss.Color("246")).
					Render("Press 'c' to mark as completed")
			}
		}
	}

	if details != "" {
		// Add a subtle box around the details
		return lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Render(details)
	}

	return ""
}

// Custom messages
type (
	fetchGardensMsg []models.Garden
	fetchBedsMsg    []models.Bed
	fetchTasksMsg   []models.Task
	fetchErrMsg     error
	completeTaskMsg struct{}
)

// Commands
func (m Model) fetchGardens() tea.Cmd {
	return func() tea.Msg {
		gardens, err := m.client.GetGardens()
		if err != nil {
			return fetchErrMsg(err)
		}
		return fetchGardensMsg(gardens)
	}
}

func (m Model) fetchBeds(gardenID string) tea.Cmd {
	return func() tea.Msg {
		beds, err := m.client.GetBeds(gardenID)
		if err != nil {
			return fetchErrMsg(err)
		}
		return fetchBedsMsg(beds)
	}
}

func (m Model) fetchTasks(gardenID, bedID string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := m.client.GetTasks(gardenID, bedID)
		if err != nil {
			return fetchErrMsg(err)
		}
		return fetchTasksMsg(tasks)
	}
}

func (m Model) completeTask(taskID string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.CompleteTask(taskID)
		if err != nil {
			return fetchErrMsg(err)
		}
		return completeTaskMsg{}
	}
}

// updateGardenTable updates the garden table data.
func (m *Model) updateGardenTable() {
	rows := []table.Row{}
	for _, g := range m.gardens {
		rows = append(rows, table.Row{
			g.ID,
			g.Name,
			g.Location,
			g.Description,
		})
	}
	m.gardenTable.SetRows(rows)
}

// updateBedTable updates the bed table data.
func (m *Model) updateBedTable() {
	rows := []table.Row{}
	for _, b := range m.beds {
		rows = append(rows, table.Row{
			b.ID,
			b.Name,
			b.Type,
			b.Size,
			b.SoilType,
		})
	}
	m.bedTable.SetRows(rows)
}

// updateTaskTable updates the task table data.
func (m *Model) updateTaskTable() {
	rows := []table.Row{}
	for _, t := range m.tasks {
		rows = append(rows, table.Row{
			t.ID,
			t.Description,
			t.DueDate.String(),
			t.Status,
			t.Priority,
		})
	}
	m.taskTable.SetRows(rows)
}

func teaProgram() tea.Model {
	m := NewModel()

	// Set reasonable initial window dimensions for local mode
	// These will be updated when the terminal sends a WindowSizeMsg
	m.width = 80
	m.height = 24

	return m
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		// Handle non-pty sessions if necessary.
		return nil, nil
	}

	width := 80
	height := 24
	if active {
		width = pty.Window.Width
		height = pty.Window.Height
	}

	m := NewModel()
	m.width = width
	m.height = height

	// Base options
	opts := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		// Crucially, wire the session's I/O to Bubble Tea
		tea.WithInput(s),
		tea.WithOutput(s),
	}

	return m, opts
}

func main() {
	// Set up SSH server
	// Enable flag for local development
	local := flag.Bool("local", false, "Run in local mode")
	flag.Parse()

	if *local {
		fmt.Println("Running Plantastic in local mode")
		fmt.Println("Press 'q' to quit, '?' for help")
		fmt.Println("Connecting to API at http://localhost:8080...")

		// Create a new program with the BubbleTea model
		p := tea.NewProgram(
			teaProgram(),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		// Run the program
		if _, err := p.Run(); err != nil {
			log.Fatalf("Error running program: %v", err)
		}
	} else {
		s, err := wish.NewServer(
			wish.WithAddress(":2222"),
			wish.WithHostKeyPath(".ssh/term_info_ed25519"),
			wish.WithBannerHandler(func(ctx ssh.Context) string {
				return "Welcome to Plantastic"
			}),
			wish.WithMiddleware(
				bm.Middleware(teaHandler),
				lm.Middleware(),
			),
		)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Starting Plantastic SSH server on port 2222...")
		fmt.Println("Connect with: ssh localhost -p 2222")

		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			if err = s.ListenAndServe(); err != nil {
				log.Fatalln(err)
			}
		}()

		<-done
		log.Println("Stopping SSH server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			log.Fatalln(err)
		}
	}
}
