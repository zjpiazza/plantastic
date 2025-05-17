package main

// An example Bubble Tea server. This will put an ssh session into alt screen
// and continually print up to date terminal information.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"

	clerkSDK "github.com/clerk/clerk-sdk-go/v2"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/zjpiazza/plantastic/cmd/tui/components"
	"github.com/zjpiazza/plantastic/internal/models"
)

const (
	host = "localhost"
	port = "23236"
)

// UI States
const (
	stateAuth = iota // New state for authentication
	stateSplash
	stateLoading
	stateReady
)

// Add these constants for auth states
const (
	authStateNone = iota
	authStateWaitingForCode
	authStateWaitingForAuth
	authStateAuthenticated
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			MarginBottom(1)

	tabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 3)

	activeTabStyle = tabStyle.Copy().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Underline(true).
			Bold(true)

	tabGap = tabStyle.Copy().
		Foreground(lipgloss.Color("236")).
		Background(lipgloss.Color("236")).
		PaddingLeft(1).
		PaddingRight(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#25A065")).
			Background(lipgloss.Color("0")).
			Bold(true).
			Align(lipgloss.Center)

	splashNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#25A065")).
			Bold(true).
			Background(lipgloss.Color("0")).
			MarginTop(1).
			MarginBottom(1).
			Align(lipgloss.Center)

	splashPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#25A065")).
				Padding(0, 1).
				MarginTop(2).
				Align(lipgloss.Center)
)

// Get splash text
func getSplashArt(width int) string {
	return "PLANTASTIC"
}

// KeyMap defines key bindings for the application
type keyMap struct {
	Tab        key.Binding
	Enter      key.Binding
	New        key.Binding
	Delete     key.Binding
	Edit       key.Binding
	Quit       key.Binding
	Help       key.Binding
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	ToggleTabs key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Tab, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Tab, k.ToggleTabs},
		{k.New, k.Delete, k.Edit, k.Enter},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch tab"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select/confirm"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new item"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete item"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit item"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
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
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	ToggleTabs: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle tabs"),
	),
}

func main() {
	err := godotenv.Load()
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			plantasticBubbleteaMiddleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

// plantasticBubbleteaMiddleware creates a custom bubbletea middleware
func plantasticBubbleteaMiddleware() wish.Middleware {
	teaHandler := func(s ssh.Session) *tea.Program {
		pty, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}
		m := initialModel(pty.Term, pty.Window.Width, pty.Window.Height)
		return tea.NewProgram(m, append(bubbletea.MakeOptions(s), tea.WithAltScreen())...)
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

type tabContent int

const (
	dashboardTab tabContent = iota
	gardensTab
	bedsTab
	tasksTab
	settingsTab
)

type model struct {
	tabTitles  []string
	activeTab  tabContent
	help       help.Model
	showHelp   bool
	width      int
	height     int
	term       string
	gardenList list.Model
	bedList    list.Model
	taskTable  table.Model
	loading    bool
	err        error

	// State management
	uiState     int
	spinner     spinner.Model
	loadingStep int
	loadMsg     string

	// Form management
	showingForm    bool
	gardenForm     components.GardenForm
	bedForm        components.BedForm
	taskForm       components.TaskForm
	activeFormType FormType

	// Storage
	storage *MemoryStorage

	// Clerk Authentication - no client stored if using global SetKey
	isAuthenticated  bool
	authTokenInput   textinput.Model
	authErrorMessage string
	// sessionClaims *clerkSession.Claims // Optionally store claims if needed later

	// Auth state management
	authState        int
	deviceCode       string // This will store the UserCode from the API
	deviceID         string // Unique ID for this TUI instance
	authError        string
	pollTicker       *time.Ticker
	verificationURI  string        // To store the URI provided by the API
	authPollInterval time.Duration // To store the polling interval from API
}

func initialModel(term string, width, height int) model {
	// Initialize Clerk Client by setting the global key
	secretKey := os.Getenv("CLERK_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("CLERK_SECRET_KEY environment variable not set.")
	}
	clerkSDK.SetKey(secretKey) // Set the key globally

	// Auth Token Input
	tokenInput := textinput.New()
	tokenInput.Placeholder = "Paste your Clerk Session Token here"
	tokenInput.Focus()
	tokenInput.CharLimit = 0 // No limit, tokens can be long
	tokenInput.Width = 50

	// Set up help
	h := help.New()
	h.ShowAll = true

	// Create a shared storage
	storage := NewMemoryStorage()

	// Load sample data
	loadSampleData(storage)

	// Create garden list
	gardenDelegate := NewGardenDelegate(list.NewDefaultItemStyles())
	gardenList := list.New([]list.Item{}, gardenDelegate, width/2, height-10)
	gardenList.Title = "Gardens"
	gardenList.SetShowHelp(false)
	gardenList.Styles.Title = titleStyle

	bedDelegate := NewBedDelegate(list.NewDefaultItemStyles())
	bedList := list.New([]list.Item{}, bedDelegate, width/2, height-10)
	bedList.Title = "Beds"
	bedList.SetShowHelp(false)
	bedList.Styles.Title = titleStyle

	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "Description", Width: 30},
		{Title: "Due Date", Width: 15},
		{Title: "Status", Width: 15},
		{Title: "Priority", Width: 10},
	}
	rows := []table.Row{}
	taskTable := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height-15),
	)
	taskTable.SetStyles(table.DefaultStyles())

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	tabNames := []string{"Dashboard", "Gardens", "Beds", "Tasks", "Settings"}

	// Generate a unique DeviceID for this TUI instance
	instanceDeviceID := uuid.New().String()

	m := model{
		tabTitles:        tabNames,
		activeTab:        dashboardTab,
		help:             h,
		term:             term,
		width:            width,
		height:           height,
		gardenList:       gardenList,
		bedList:          bedList,
		taskTable:        taskTable,
		uiState:          stateAuth, // Start with authentication state
		spinner:          s,
		loadingStep:      0,
		loadMsg:          "Initializing Plantastic...",
		storage:          storage,
		authTokenInput:   tokenInput,
		deviceID:         instanceDeviceID, // Set the generated DeviceID
		authPollInterval: 5 * time.Second,  // Default poll interval
	}

	m.refreshGardenList() // Keep this, or move it after successful auth if data depends on user

	return m
}

// loadSampleData adds sample gardens, beds, and tasks to the storage
func loadSampleData(storage *MemoryStorage) {
	// Create sample gardens
	garden1 := models.NewGarden("Backyard Garden", "Behind the house", "Main vegetable and herb garden")
	garden2 := models.NewGarden("Front Garden", "Front yard", "Ornamental flowers and shrubs")
	garden3 := models.NewGarden("Container Garden", "Patio", "Container plants for small spaces")

	storage.AddGarden(garden1)
	storage.AddGarden(garden2)
	storage.AddGarden(garden3)

	// Create sample beds for Backyard Garden
	bed1 := models.NewBed(garden1.ID, "Tomato Bed", "Raised", "4' x 8'", "Loamy soil mix", "Various tomato varieties")
	bed2 := models.NewBed(garden1.ID, "Herb Bed", "In-ground", "3' x 6'", "Sandy loam", "Basil, thyme, oregano, and rosemary")
	bed3 := models.NewBed(garden1.ID, "Greens Bed", "Raised", "4' x 4'", "Compost-rich mix", "Lettuce, spinach, kale")

	storage.AddBed(bed1)
	storage.AddBed(bed2)
	storage.AddBed(bed3)

	// Create sample beds for Front Garden
	bed4 := models.NewBed(garden2.ID, "Rose Bed", "In-ground", "6' x 3'", "Rose soil mix", "Red and pink roses")
	bed5 := models.NewBed(garden2.ID, "Tulip Border", "In-ground", "8' x 2'", "Bulb soil mix", "Spring tulips and daffodils")

	storage.AddBed(bed4)
	storage.AddBed(bed5)

	// Create sample tasks
	now := time.Now()

	// Tasks for Backyard Garden - Tomato Bed
	task1 := models.NewTask(garden1.ID, &bed1.ID, "Water tomatoes", now.AddDate(0, 0, 1), models.TaskStatusPending, models.PriorityHigh)
	task2 := models.NewTask(garden1.ID, &bed1.ID, "Add fertilizer", now.AddDate(0, 0, 7), models.TaskStatusPending, models.PriorityMedium)
	task3 := models.NewTask(garden1.ID, &bed1.ID, "Check for pests", now.AddDate(0, 0, 3), models.TaskStatusPending, models.PriorityLow)

	storage.AddTask(task1)
	storage.AddTask(task2)
	storage.AddTask(task3)

	// Tasks for Backyard Garden - Herb Bed
	task4 := models.NewTask(garden1.ID, &bed2.ID, "Harvest basil", now.AddDate(0, 0, 2), models.TaskStatusPending, models.PriorityMedium)
	task5 := models.NewTask(garden1.ID, &bed2.ID, "Prune rosemary", now.AddDate(0, 0, 14), models.TaskStatusPending, models.PriorityLow)

	storage.AddTask(task4)
	storage.AddTask(task5)

	// Tasks for Front Garden
	task6 := models.NewTask(garden2.ID, nil, "Mulch all beds", now.AddDate(0, 0, 5), models.TaskStatusPending, models.PriorityHigh)
	task7 := models.NewTask(garden2.ID, &bed4.ID, "Spray roses for pests", now.AddDate(0, 0, 4), models.TaskStatusPending, models.PriorityMedium)

	storage.AddTask(task6)
	storage.AddTask(task7)
}

type item struct {
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadingTick() tea.Cmd {
	return tea.Tick(350*time.Millisecond, func(t time.Time) tea.Msg {
		return loadingTickMsg(t)
	})
}

type loadingTickMsg time.Time

func (m model) Init() tea.Cmd {
	if m.uiState == stateAuth {
		m.authState = authStateNone
		return tea.Batch(
			m.spinner.Tick,
			tea.Tick(time.Millisecond, func(t time.Time) tea.Msg {
				return tickMsg(t)
			}),
		)
	}
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle auth state
	if m.uiState == stateAuth {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "Q", "ctrl+c":
				return m, tea.Quit
			}

		case tickMsg:
			if m.authState == authStateNone {
				m.authState = authStateWaitingForCode
				return m, m.requestDeviceCode()
			}
			return m, m.spinner.Tick

		case deviceCodeMsg:
			m.deviceCode = msg.code
			if msg.verificationURI != "" {
				m.verificationURI = msg.verificationURI
			}
			if msg.interval > 0 {
				m.authPollInterval = time.Duration(msg.interval) * time.Second
			}

			m.authState = authStateWaitingForAuth
			// Start polling for auth status
			if m.pollTicker != nil {
				m.pollTicker.Stop()
			}
			pollDuration := m.authPollInterval
			if pollDuration < 2*time.Second {
				pollDuration = 2 * time.Second // Minimum polling interval
			}
			m.pollTicker = time.NewTicker(pollDuration)
			return m, tea.Tick(pollDuration, func(t time.Time) tea.Msg {
				return authPendingMsg{}
			})

		case authPendingMsg:
			return m, m.checkAuthStatus()

		case authSuccessMsg:
			log.Info("Authentication successful, received token", "length", len(msg.token))
			if m.pollTicker != nil {
				m.pollTicker.Stop()
			}
			// Store the token and proceed
			m.isAuthenticated = true
			m.uiState = stateSplash
			return m, tea.Batch(m.spinner.Tick, loadingTick())

		case errMsg:
			m.authError = msg.error.Error()
			log.Error("Auth error", "error", m.authError)
			return m, nil
		}
	}

	// Handle form updates if we're showing a form
	if m.showingForm {
		switch m.activeFormType {
		case FormTypeGarden:
			formModel, cmd := m.gardenForm.Update(msg)
			m.gardenForm = formModel.(components.GardenForm)
			cmds = append(cmds, cmd)

			if m.gardenForm.Submitted() || m.gardenForm.Cancelled() {
				m.showingForm = false
				if m.gardenForm.Submitted() {
					m.refreshGardenList()
				}
			}
			return m, tea.Batch(cmds...)

		case FormTypeBed:
			formModel, cmd := m.bedForm.Update(msg)
			m.bedForm = formModel.(components.BedForm)
			cmds = append(cmds, cmd)

			if m.bedForm.Submitted() || m.bedForm.Cancelled() {
				m.showingForm = false
				if m.bedForm.Submitted() {
					if garden, ok := m.getSelectedGarden(); ok {
						m.refreshBedList(garden.ID)
					}
				}
			}
			return m, tea.Batch(cmds...)

		case FormTypeTask:
			formModel, cmd := m.taskForm.Update(msg)
			m.taskForm = formModel.(components.TaskForm)
			cmds = append(cmds, cmd)

			if m.taskForm.Submitted() || m.taskForm.Cancelled() {
				m.showingForm = false
				if m.taskForm.Submitted() {
					var gardenID string
					var bedID *string
					if garden, ok := m.getSelectedGarden(); ok {
						gardenID = garden.ID
						if bed, ok := m.getSelectedBed(); ok {
							bedID = &bed.ID
						}
						m.refreshTaskTable(gardenID, bedID)
					}
				}
			}
			return m, tea.Batch(cmds...)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle Enter key on splash screen to proceed to loading screen
		if m.uiState == stateSplash && key.Matches(msg, keys.Enter) {
			m.uiState = stateLoading
			return m, tea.Batch(
				m.spinner.Tick,
				loadingTick(),
			)
		}

		// Only accept quit command during splash and loading
		if (m.uiState == stateSplash || m.uiState == stateLoading) && key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}

		// Handle other keys only when app is fully loaded
		if m.uiState == stateReady {
			switch {
			case key.Matches(msg, keys.Quit):
				return m, tea.Quit

			case key.Matches(msg, keys.Help):
				m.showHelp = !m.showHelp

			case key.Matches(msg, keys.Tab):
				m.activeTab = (m.activeTab + 1) % tabContent(len(m.tabTitles))

			case key.Matches(msg, keys.ToggleTabs):
				// You could toggle tabs/navigation view here if needed

			case key.Matches(msg, keys.New):
				switch m.activeTab {
				case gardensTab:
					m.gardenForm = components.NewGardenForm(m, m.width, m.height, func(garden models.Garden) {
						m.refreshGardenList()
					})
					m.activeFormType = FormTypeGarden
					m.showingForm = true

				case bedsTab:
					if selectedGarden, ok := m.getSelectedGarden(); ok {
						m.bedForm = components.NewBedForm(m, selectedGarden.ID, m.width, m.height, func(bed models.Bed) {
							m.refreshBedList(bed.GardenID)
						})
						m.activeFormType = FormTypeBed
						m.showingForm = true
					}

				case tasksTab:
					if selectedGarden, ok := m.getSelectedGarden(); ok {
						var bedID *string
						if selectedBed, ok := m.getSelectedBed(); ok {
							bedID = &selectedBed.ID
						}
						m.taskForm = components.NewTaskForm(m, selectedGarden.ID, bedID, m.width, m.height, func(task models.Task) {
							m.refreshTaskTable(task.GardenID, task.BedID)
						})
						m.activeFormType = FormTypeTask
						m.showingForm = true
					}
				}

			case key.Matches(msg, keys.Edit):
				switch m.activeTab {
				case gardensTab:
					if selectedGarden, ok := m.getSelectedGarden(); ok {
						m.gardenForm = components.NewGardenForm(m, m.width, m.height, func(garden models.Garden) {
							m.refreshGardenList()
						})
						m.gardenForm.SetGarden(selectedGarden)
						m.activeFormType = FormTypeGarden
						m.showingForm = true
					}

				case bedsTab:
					if selectedBed, ok := m.getSelectedBed(); ok {
						m.bedForm = components.NewBedForm(m, selectedBed.GardenID, m.width, m.height, func(bed models.Bed) {
							m.refreshBedList(bed.GardenID)
						})
						m.bedForm.SetBed(selectedBed)
						m.activeFormType = FormTypeBed
						m.showingForm = true
					}

				case tasksTab:
					if selectedTask, ok := m.getSelectedTask(); ok {
						m.taskForm = components.NewTaskForm(m, selectedTask.GardenID, selectedTask.BedID, m.width, m.height, func(task models.Task) {
							m.refreshTaskTable(task.GardenID, task.BedID)
						})
						m.taskForm.SetTask(selectedTask)
						m.activeFormType = FormTypeTask
						m.showingForm = true
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = m.width
		// Update list dimensions
		m.gardenList.SetSize(m.width/2, m.height-10)
		m.bedList.SetSize(m.width/2, m.height-10)
		// Update table height
		m.taskTable.SetHeight(m.height - 15)
		// Update auth input width
		m.authTokenInput.Width = m.width - 20 // Adjust as needed

	case loadingTickMsg:
		if m.uiState == stateLoading {
			m.loadingStep++
			switch m.loadingStep {
			case 1:
				m.loadMsg = "Connecting to garden database..."
			case 2:
				m.loadMsg = "Loading garden data..."
			case 3:
				m.loadMsg = "Loading bed information..."
			case 4:
				m.loadMsg = "Checking task schedules..."
			case 5:
				m.loadMsg = "Preparing user interface..."
			case 6:
				m.uiState = stateReady
				return m, tick()
			}
			cmds = append(cmds, loadingTick())
		}

	case spinner.TickMsg:
		if m.uiState == stateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tickMsg:
		cmds = append(cmds, tick())
	}

	// Handle tab-specific updates
	switch m.activeTab {
	case gardensTab:
		var cmd tea.Cmd
		m.gardenList, cmd = m.gardenList.Update(msg)
		cmds = append(cmds, cmd)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, keys.Enter) {
			if selectedGarden, ok := m.getSelectedGarden(); ok {
				m.refreshBedList(selectedGarden.ID)
				m.refreshTaskTable(selectedGarden.ID, nil)
			}
		}

	case bedsTab:
		var cmd tea.Cmd
		m.bedList, cmd = m.bedList.Update(msg)
		cmds = append(cmds, cmd)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, keys.Enter) {
			if selectedBed, ok := m.getSelectedBed(); ok {
				gardenID := selectedBed.GardenID
				bedID := selectedBed.ID
				m.refreshTaskTable(gardenID, &bedID)
			}
		}

	case tasksTab:
		var cmd tea.Cmd
		m.taskTable, cmd = m.taskTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// renderSplash renders the splash screen
func (m model) renderSplash() string {
	s := strings.Builder{}

	// Add some padding at the top based on terminal height
	padding := m.height / 3
	if padding > 0 {
		s.WriteString(strings.Repeat("\n", padding))
	}

	// Add app name in styled text
	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#25A065")).
		Bold(true).
		Width(m.width).
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#25A065")).
		Padding(1, 3)

	appName := nameStyle.Render("PLANTASTIC")
	s.WriteString(appName)
	s.WriteString("\n\n")

	// Add subtitle
	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Width(m.width).
		Align(lipgloss.Center)

	subtitle := subtitleStyle.Render("Your Plant Management Terminal Interface")
	s.WriteString(subtitle)
	s.WriteString("\n\n")

	// Add a prompt to press Enter to continue
	prompt := splashPromptStyle.Width(m.width).Render("Press ENTER to continue")
	s.WriteString(prompt)

	return s.String()
}

// renderTabs renders the tab bar
func (m model) renderTabs() string {
	var renderedTabs []string

	for i, tab := range m.tabTitles {
		if i == int(m.activeTab) {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, tabStyle.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (m model) renderLoading() string {
	// Center the loading content
	s := strings.Builder{}

	// Add some padding at the top
	s.WriteString(strings.Repeat("\n", m.height/4))

	// Add welcome message
	welcome := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#25A065")).
		Bold(true).
		Width(m.width).
		Align(lipgloss.Center).
		Render("Welcome to Plantastic!")

	s.WriteString(welcome)
	s.WriteString("\n\n")

	// Add spinner and loading message
	loadingText := fmt.Sprintf("%s %s", m.spinner.View(), m.loadMsg)
	loading := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(loadingText)

	s.WriteString(loading)
	s.WriteString("\n\n")

	// Add a hint to press q to quit
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(m.width).
		Align(lipgloss.Center).
		Render("Press 'q' to quit")

	s.WriteString(hint)

	return s.String()
}

func (m model) View() string {
	// Show auth screen first
	if m.uiState == stateAuth {
		return m.renderAuthScreen()
	}

	// Show form if one is active
	if m.showingForm {
		switch m.activeFormType {
		case FormTypeGarden:
			return m.gardenForm.View()
		case FormTypeBed:
			return m.bedForm.View()
		case FormTypeTask:
			return m.taskForm.View()
		}
	}

	// Show splash screen first
	if m.uiState == stateSplash {
		return m.renderSplash()
	}

	// Then show loading screen
	if m.uiState == stateLoading {
		return m.renderLoading()
	}

	// When ready, show the main UI
	var content string

	// Render tabs
	tabsView := m.renderTabs()

	// Render main content based on active tab
	switch m.activeTab {
	case dashboardTab:
		content = m.renderDashboard()
	case gardensTab:
		content = m.gardenList.View()
	case bedsTab:
		content = m.bedList.View()
	case tasksTab:
		content = m.renderTasks()
	case settingsTab:
		content = m.renderSettings()
	}

	// Render help
	helpView := ""
	if m.showHelp {
		helpView = m.help.View(keys)
	} else {
		helpView = helpStyle.Render("Press ? for help")
	}

	// Put it all together
	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s",
		titleStyle.Render("PLANTASTIC - Your Plant Management TUI"),
		tabsView,
		content,
		helpView,
	)
}

func (m model) renderAuthScreen() string {
	var b strings.Builder

	title := splashNameStyle.Width(m.width).Render("Plantastic Authentication")
	b.WriteString(strings.Repeat("\n", m.height/4))
	b.WriteString(title)
	b.WriteString("\n\n")

	switch m.authState {
	case authStateNone:
		// Initial state - show welcome message
		msg := "Initializing device authentication..."
		b.WriteString(lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(msg))

	case authStateWaitingForCode:
		// Waiting for API response
		msg := m.spinner.View() + " Requesting device code..."
		b.WriteString(lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(msg))

	case authStateWaitingForAuth:
		// Show the code and URL
		linkURL := m.verificationURI
		if linkURL == "" {
			// Fallback if verification URI wasn't received from API
			linkURL = fmt.Sprintf("http://localhost:8080/link/%s", m.deviceCode)
		}

		codeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#25A065"))
		urlStyle := lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("#4B9CD3"))

		msg := fmt.Sprintf("Please visit:\n\n%s\n\nAnd enter this code: %s\n\nWaiting for authentication...",
			urlStyle.Render(linkURL),
			codeStyle.Render(m.deviceCode))
		b.WriteString(lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(msg))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(m.spinner.View()))
	}

	if m.authError != "" {
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("9")).
			Render(m.authError))
	}

	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Bold(true).
		Render("Press Q to quit"))

	return b.String()
}

func (m model) renderDashboard() string {
	// Define the number of panels and the gap between them
	numPanels := 3
	gap := 2 // Number of spaces for the gap

	// Calculate the total width available for all panels, excluding gaps between them
	totalContentWidth := m.width - ((numPanels - 1) * gap)

	// Calculate the width for each individual panel (container)
	// This is the width the panelStyle (outer box) should occupy
	panelOuterWidth := totalContentWidth / numPanels

	// Define panel padding and border size to calculate inner content width
	panelPadding := 1 // Padding on each side inside the border
	panelBorder := 1  // Border on each side
	panelInnerWidth := panelOuterWidth - (panelPadding * 2) - (panelBorder * 2)

	// Create content for each panel
	upcomingContent := []string{
		"Water tomatoes (in 2 days)",
		"Harvest basil (in 3 days)",
		"Plant new flowers (in 7 days)",
	}

	statsContent := []string{
		"3 Gardens",
		"4 Active beds",
		"7 Total tasks (3 pending)",
	}

	weatherContent := []string{
		"Partly Cloudy, 75°F",
		"Precipitation: 20% chance",
		"", // Add an empty line to match height if needed
	}

	// Style for panel headers
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#25A065")).
		Bold(true).
		Width(panelInnerWidth). // Header uses inner width
		Align(lipgloss.Center).
		MarginBottom(1).
		Underline(true)

	// Style for panel content text
	textStyle := lipgloss.NewStyle().
		Width(panelInnerWidth). // Content text uses inner width
		PaddingLeft(1)          // Indent content slightly

	// Style for the panel container (the box)
	panelBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#25A065")).
		Width(panelOuterWidth). // This is the crucial width for the outer box
		Padding(panelPadding)
		// MarginRight will be handled by lipgloss.JoinHorizontal with a gap string

	// Helper function to format items with an arrow
	formatItems := func(items []string) string {
		var b strings.Builder
		for _, item := range items {
			if item != "" {
				b.WriteString("-> " + item + "\n")
			} else {
				b.WriteString("\n") // Preserve empty lines for height matching
			}
		}
		return b.String()
	}

	// Create each panel
	upcomingPanel := panelBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			headerStyle.Render("UPCOMING TASKS"),
			textStyle.Render(formatItems(upcomingContent)),
		),
	)

	statsPanel := panelBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			headerStyle.Render("GARDEN STATS"),
			textStyle.Render(formatItems(statsContent)),
		),
	)

	weatherPanel := panelBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			headerStyle.Render("WEATHER"),
			textStyle.Render(formatItems(weatherContent)),
		),
	)

	// For narrow terminals, stack them vertically
	// Threshold can be adjusted based on desired switch point
	if m.width < (panelOuterWidth*numPanels + (numPanels-1)*gap + 5) { // +5 for some buffer
		return lipgloss.JoinVertical(lipgloss.Center,
			upcomingPanel,
			statsPanel,
			weatherPanel,
		)
	}

	// For wider terminals, join horizontally with a defined gap string
	gapStr := strings.Repeat(" ", gap)
	return lipgloss.JoinHorizontal(lipgloss.Top,
		upcomingPanel,
		gapStr,
		statsPanel,
		gapStr,
		weatherPanel,
	)
}

func (m model) renderTasks() string {
	return fmt.Sprintf("%s\n\nPress Enter to view task details. Press 'n' to create a new task.", m.taskTable.View())
}

func (m model) renderSettings() string {
	return `SETTINGS

-> User Preferences
-> API Connections
-> Theme Settings
-> Export Data
-> About

Use arrow keys to navigate and Enter to select.`
}

// Implement the Storage interface for GardenForm
func (m model) GetGarden(id string) (models.Garden, bool) {
	return m.storage.GetGarden(id)
}

func (m model) AddGarden(garden models.Garden) error {
	return m.storage.AddGarden(garden)
}

func (m model) UpdateGarden(garden models.Garden) error {
	return m.storage.UpdateGarden(garden)
}

// Implement the BedStorage interface for BedForm
func (m model) GetBed(id string) (models.Bed, bool) {
	return m.storage.GetBed(id)
}

func (m model) AddBed(bed models.Bed) error {
	return m.storage.AddBed(bed)
}

func (m model) UpdateBed(bed models.Bed) error {
	return m.storage.UpdateBed(bed)
}

// Helper methods to get selected items from the lists/tables
func (m model) getSelectedGarden() (models.Garden, bool) {
	// Check if we have any gardens in the list
	if len(m.gardenList.Items()) == 0 {
		return models.Garden{}, false
	}

	// Get the currently selected garden
	index := m.gardenList.Index()
	if index < 0 || index >= len(m.gardenList.Items()) {
		return models.Garden{}, false
	}

	// Convert to GardenItem
	item := m.gardenList.Items()[index]
	if gardenItem, ok := item.(GardenItem); ok {
		return gardenItem.garden, true
	}

	return models.Garden{}, false
}

func (m model) getSelectedBed() (models.Bed, bool) {
	// Check if we have any beds in the list
	if len(m.bedList.Items()) == 0 {
		return models.Bed{}, false
	}

	// Get the currently selected bed
	index := m.bedList.Index()
	if index < 0 || index >= len(m.bedList.Items()) {
		return models.Bed{}, false
	}

	// Convert to BedItem
	item := m.bedList.Items()[index]
	if bedItem, ok := item.(BedItem); ok {
		return bedItem.bed, true
	}

	return models.Bed{}, false
}

func (m model) getSelectedTask() (models.Task, bool) {
	// Get the currently selected task from the table
	if m.taskTable.Cursor() >= len(m.taskTable.Rows()) {
		return models.Task{}, false
	}

	row := m.taskTable.SelectedRow()
	if len(row) == 0 {
		return models.Task{}, false
	}

	// Get task ID from the first column
	taskID := row[0]

	// Look up the task in the storage
	return m.storage.GetTask(taskID)
}

// Implement the TaskStorage interface for TaskForm
func (m model) GetTask(id string) (models.Task, bool) {
	return m.storage.GetTask(id)
}

func (m model) AddTask(task models.Task) error {
	return m.storage.AddTask(task)
}

func (m model) UpdateTask(task models.Task) error {
	return m.storage.UpdateTask(task)
}

// Add refresh methods to update the UI lists and tables with data from storage
func (m *model) refreshGardenList() {
	gardens := m.storage.GetGardens()
	items := make([]list.Item, len(gardens))
	for i, garden := range gardens {
		items[i] = GardenItem{garden: garden}
	}
	m.gardenList.SetItems(items)
}

func (m *model) refreshBedList(gardenID string) {
	beds := m.storage.GetBeds(gardenID)
	items := make([]list.Item, len(beds))
	for i, bed := range beds {
		items[i] = BedItem{bed: bed}
	}
	m.bedList.SetItems(items)
}

func (m *model) refreshTaskTable(gardenID string, bedID *string) {
	tasks := m.storage.GetTasks(gardenID, bedID)
	rows := make([]table.Row, len(tasks))
	for i, task := range tasks {
		dueDate := task.DueDate.Format("2006-01-02")
		rows[i] = table.Row{
			task.ID,
			task.Description,
			dueDate,
			task.Status,
			task.Priority,
		}
	}
	m.taskTable.SetRows(rows)
}

// Add these new functions for device auth
func (m model) requestDeviceCode() tea.Cmd {
	return func() tea.Msg {
		// Ensure deviceID is set
		if m.deviceID == "" {
			return errMsg{fmt.Errorf("TUI device ID is not set")}
		}

		log.Info("Requesting device code", "deviceID", m.deviceID)

		payload := map[string]string{"device_id": m.deviceID}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return errMsg{fmt.Errorf("failed to marshal request payload: %w", err)}
		}

		req, err := http.NewRequest("POST", "http://localhost:8000/device/request-code", bytes.NewBuffer(jsonPayload))
		if err != nil {
			return errMsg{fmt.Errorf("failed to create request: %w", err)}
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Error("Failed to request device code", "error", err)
			return errMsg{err}
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Info("Device code response", "status", resp.Status, "body", string(bodyBytes))

		if resp.StatusCode != http.StatusOK {
			return errMsg{fmt.Errorf("failed to request device code, status: %s, body: %s", resp.Status, string(bodyBytes))}
		}

		var result struct {
			UserCode        string `json:"user_code"`
			VerificationURI string `json:"verification_uri"`
			ExpiresIn       int    `json:"expires_in"`
			Interval        int    `json:"interval"`
		}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return errMsg{fmt.Errorf("failed to decode device code response: %w", err)}
		}

		if result.UserCode == "" {
			return errMsg{fmt.Errorf("received empty user_code from API")}
		}

		log.Info("Received device code", "code", result.UserCode)

		// Return all details in the message
		return deviceCodeMsg{
			code:            result.UserCode,
			verificationURI: result.VerificationURI,
			interval:        result.Interval,
			expiresIn:       result.ExpiresIn,
		}
	}
}

func (m model) checkAuthStatus() tea.Cmd {
	return func() tea.Msg {
		if m.deviceCode == "" { // This is the user_code
			return errMsg{fmt.Errorf("no device code to check status for")}
		}

		log.Info("Checking auth status", "deviceCode", m.deviceCode)

		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/device/check-status?code=%s", m.deviceCode), nil)
		if err != nil {
			return errMsg{fmt.Errorf("failed to create request: %w", err)}
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Error("Error checking auth status", "error", err)
			// Continue polling even on connection errors
			return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return authPendingMsg{}
			})
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Info("Auth status response", "status", resp.Status, "body", string(bodyBytes))

		if resp.StatusCode != http.StatusOK {
			return errMsg{fmt.Errorf("failed to check auth status, status: %s, body: %s", resp.Status, string(bodyBytes))}
		}

		// Re-use the buffer we already read
		var result struct {
			Status string `json:"status"`
			Token  string `json:"token,omitempty"`
			Error  string `json:"error,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return errMsg{fmt.Errorf("failed to decode auth status response: %w", err)}
		}

		switch result.Status {
		case "activated":
			log.Info("Device has been activated")
			if result.Token == "" {
				return errMsg{fmt.Errorf("status is activated but token is missing")}
			}
			return authSuccessMsg{token: result.Token}
		case "pending_activation":
			log.Info("Still waiting for activation")
			return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return authPendingMsg{}
			})
		case "failed":
			return errMsg{fmt.Errorf("authentication failed: %s", result.Error)}
		default:
			return errMsg{fmt.Errorf("unknown authentication status: %s", result.Status)}
		}
	}
}

// Add message types
type (
	deviceCodeMsg struct {
		code            string
		verificationURI string
		interval        int
		expiresIn       int
	}
	authSuccessMsg struct{ token string }
	authPendingMsg struct{}
	errMsg         struct{ error error }
)
