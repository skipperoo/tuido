package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/skipperoo/tuido/internal/core"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Styles ---
var (
	cMagenta = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	cBlue    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	cCyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	cGreen   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	cYellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	cRed     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	cGray    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// --- Application States ---
type appState int

const (
	stateViewMain appState = iota
	stateSelectTask
	stateEditTaskInput
	stateHistoryView
)

type selectMode int

const (
	modeDone selectMode = iota
	modeUndone
	modeEdit
	modeRemove
)

// --- Model ---
type model struct {
	state      appState
	filePath   string
	configPath string
	author     string

	// Data
	entries []core.Entry

	// Input & Viewport
	textInput textinput.Model
	viewport  viewport.Model

	// Selection Logic
	cursor        int
	selectList    []core.Entry        // Items being selected
	selectedIDs   map[string]struct{} // IDs of selected items (for multiselect)
	selectionMode selectMode          // Are we marking done, editing, or removing?

	// Messages
	msg string
}

func initialModel() model {
	// Setup text input
	ti := textinput.New()
	ti.Placeholder = "Type a note, /todo, or command..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	// Setup viewport
	vp := viewport.New(80, 20)

	// Determine CWD and File Path

cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting CWD: %v\n", err)
		os.Exit(1)
	}
	targetFile := filepath.Join(cwd, ".tuido")

	// Load Config
	cfg, _ := core.LoadConfig()
	author := cfg.Author
	if author == "" {
		// Ideally prompt user, but for now default to "User" or OS user
		userEnv := os.Getenv("USER")
		if userEnv == "" {
			author = "User"
		} else {
			author = userEnv
		}
	}

	m := model{
		state:       stateViewMain,
		filePath:    targetFile,
		author:      author,
		textInput:   ti,
		viewport:    vp,
		entries:     []core.Entry{},
		selectedIDs: make(map[string]struct{}),
	}

	m.reloadEntries()
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// --- Logic Helpers ---

func (m *model) reloadEntries() {
	entries, err := core.LoadEntries(m.filePath)
	if err != nil {
		m.msg = fmt.Sprintf("Error loading file: %v", err)
		return
	}
	m.entries = entries
	m.updateViewport()
}

func (m *model) save() {
	err := core.SaveEntries(m.filePath, m.entries)
	if err != nil {
		m.msg = fmt.Sprintf("Error saving file: %v", err)
	}
	m.reloadEntries()
}

func (m *model) updateViewport() {
	var sb strings.Builder

	// Helper to format date
	fmtDate := func(t time.Time) string {
		return t.Format("02-01-2006 15:04")
	}

	for _, e := range m.entries {
		line := ""
		if e.Type == core.TypeNote {
			// [ Author, datetime ] - COMMENT TEXT
			line = fmt.Sprintf("[ %s, %s ] - %s",
				cMagenta.Render(e.Author),
				cYellow.Render(fmtDate(e.CreatedAt)),
				e.Text)
		} else if e.Type == core.TypeTodo {
			if e.CompletedAt == nil {
				// [ TODO ] - [ Author, created_at ] - TASK TEXT
				line = fmt.Sprintf("[ %s ] - [ %s, %s ] - %s",
					cGreen.Render("TODO"),
					cMagenta.Render(e.Author),
					cYellow.Render(fmtDate(e.CreatedAt)),
					e.Text)
			} else {
				// Skip completed tasks in main view
				continue
			}
		}
		sb.WriteString(line + "\n")
	}

	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

func (m model) renderHistoryContent() string {
	var sb strings.Builder
	fmtDate := func(t time.Time) string {
		return t.Format("02-01-2006 15:04")
	}

	completed := core.GetCompletedTodos(m.entries)
	for _, e := range completed {
		line := fmt.Sprintf("[ %s ] - [ %s, %s -> %s ] - %s",
			cCyan.Render("DONE"),
			cMagenta.Render(e.Author),
			cYellow.Render(fmtDate(e.CreatedAt)),
			cYellow.Render(fmtDate(*e.CompletedAt)),
			e.Text)
		sb.WriteString(line + "\n")
	}
	return sb.String()
}

// --- Update Loop ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		// Static height: TopPadding(2) + Title(6) + Author(1) + Gap(1) + Input(1) + Help(1) = 12
		// Let's use 14 to be safe and avoid scrolling issues.
		m.viewport.Height = msg.Height - 14
		if m.viewport.Height < 1 {
			m.viewport.Height = 1
		}
	}

	switch m.state {
	case stateViewMain:
		return m.updateMain(msg)
	case stateSelectTask:
		return m.updateTaskSelect(msg)
	case stateEditTaskInput:
		return m.updateEditTask(msg)
	case stateHistoryView:
		return m.updateHistory(msg)
	}

	return m, nil
}

func (m model) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.msg = "" // Reset notification on any action

		switch msg.Type {
		case tea.KeyEnter:
			val := m.textInput.Value()
			m.textInput.SetValue("")

			// Commands
			if strings.HasPrefix(val, "/") {
				cmdStr := strings.Split(val, " ")[0]
				switch cmdStr {
				case "/exit":
					return m, tea.Quit
				case "/todo", "/t":
					text := strings.TrimPrefix(val, "/todo ")
					if text == "" {
						text = strings.TrimPrefix(val, "/t ")
					}
					if text != "" {
						m.entries = core.AddEntry(m.entries, text, m.author, core.TypeTodo)
						m.save()
					}
				case "/done", "/d":
					m.prepareTaskSelection(modeDone)
				case "/undone":
					m.prepareTaskSelection(modeUndone)
				case "/rm":
					m.prepareTaskSelection(modeRemove)
				case "/edit", "/e":
					m.prepareTaskSelection(modeEdit)
				case "/dhist":
					m.state = stateHistoryView
					m.viewport.SetContent(m.renderHistoryContent())
					m.viewport.GotoBottom()
				case "/author":
					parts := strings.Fields(val)
					if len(parts) > 1 {
						name := strings.Join(parts[1:], " ")
						m.author = name
						core.SaveConfig(core.Config{Author: name})
						m.msg = "Author updated to " + name
					}
				case "/export":
					ts := time.Now().Format("20060102_150405")
					filename := fmt.Sprintf("todo_%s.md", ts)
					content := core.GenerateExportMarkdown(m.entries)
					err := os.WriteFile(filename, []byte(content), 0o644)
					if err != nil {
						m.msg = fmt.Sprintf("Export failed: %v", err)
					} else {
						m.msg = fmt.Sprintf("Exported to %s", filename)
					}
				case "/help":
					m.msg = "Commands: /todo, /done, /undone, /rm, /edit, /dhist, /author, /export, /exit"
				}
			} else if val != "" {
				// Regular Note
				m.entries = core.AddEntry(m.entries, val, m.author, core.TypeNote)
				m.save()
			}
			return m, nil

		// Viewport Scrolling
		case tea.KeyUp, tea.KeyPgUp:
			m.viewport.ScrollUp(1)
		case tea.KeyDown, tea.KeyPgDown:
			m.viewport.ScrollDown(1)
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) updateHistory(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		// Any key returns to main
		m.state = stateViewMain
		m.updateViewport()
	}
	return m, nil
}

func (m *model) prepareTaskSelection(mode selectMode) {
	m.selectList = []core.Entry{}
	m.selectionMode = mode
	m.selectedIDs = make(map[string]struct{})

	switch mode {
	case modeDone:
		m.selectList = core.GetActiveTodos(m.entries)
	case modeUndone:
		m.selectList = core.GetCompletedTodos(m.entries)
	case modeRemove, modeEdit:
		m.selectList = core.GetActiveItems(m.entries)
	}

	if len(m.selectList) == 0 {
		m.msg = "Nothing to select!"
		return
	}
	m.state = stateSelectTask
	m.cursor = 0
}

func (m model) updateTaskSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.selectList)-1 {
				m.cursor++
			}
		case " ":
			if m.selectionMode == modeRemove {
				id := m.selectList[m.cursor].ID
				if _, selected := m.selectedIDs[id]; selected {
					delete(m.selectedIDs, id)
				} else {
					m.selectedIDs[id] = struct{}{}
				}
			}
		case "enter":
			if m.selectionMode == modeRemove {
				// If nothing is selected via space, remove the item under cursor
				if len(m.selectedIDs) == 0 {
					m.entries = core.RemoveEntry(m.entries, m.selectList[m.cursor].ID)
				} else {
					m.entries = core.RemoveEntries(m.entries, m.selectedIDs)
				}
			} else {
				selected := m.selectList[m.cursor]
				if m.selectionMode == modeDone {
					m.entries = core.MarkDone(m.entries, selected.ID)
				} else if m.selectionMode == modeUndone {
					m.entries = core.MarkUndone(m.entries, selected.ID)
				} else if m.selectionMode == modeEdit {
					m.state = stateEditTaskInput
					m.textInput.SetValue(selected.Text)
					m.textInput.Focus()
					return m, nil
				}
			}

			m.save()
			m.state = stateViewMain

		case "esc":
			m.state = stateViewMain
		}
	}
	return m, nil
}

func (m model) updateEditTask(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Save edit
			selected := m.selectList[m.cursor]
			m.entries = core.EditEntry(m.entries, selected.ID, m.textInput.Value())
			m.save()
			m.textInput.SetValue("")
			m.state = stateViewMain
			return m, nil
		case tea.KeyEsc:
			m.textInput.SetValue("")
			m.state = stateViewMain
			return m, nil
		}
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// --- View ---

func (m model) View() string {
	switch m.state {
	case stateSelectTask:
		return m.viewTaskSelect()
	case stateEditTaskInput:
		return fmt.Sprintf("\n%s\n\n%s\n\n(Esc to cancel)", cMagenta.Render("Edit Task:"), m.textInput.View())
	case stateHistoryView:
		return fmt.Sprintf("%s\n\n%s\n\n%s",
			cMagenta.Render("History (Completed Tasks)"),
			m.viewport.View(),
			cGray.Render("Press any key to go back"))
	default:
		return m.viewMain()
	}
}

func (m model) viewMain() string {
	title := cMagenta.Render(`
	_______    _ _____        
	|__   __|  (_)  __ \       
	   | |_   _ _| |  | | ___  
	   | | | | | | |  | |/ _ \ 
	   | | |_| | | |__| | (_) |
	   |_|\__,_|_|_____/ \___/ `)

	header := fmt.Sprintf("%s\nAuthor: %s", title, cBlue.Render(m.author))
	help := cGray.Render(" Type to add note | /todo [text] | /help | /exit")

	if m.msg != "" {
		help = cRed.Render(m.msg) + "\n" + help
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		"\n",
		header,
		"\n",
		m.viewport.View(),
		"\n",
		m.textInput.View(),
		help,
	)
}

func (m model) viewTaskSelect() string {
	title := "Select Item"
	switch m.selectionMode {
	case modeDone:
		title = "Mark as Done"
	case modeUndone:
		title = "Mark as Undone"
	case modeRemove:
		title = "Remove Item"
	case modeEdit:
		title = "Edit Item"
	}

	ss := cMagenta.Render(title) + "\n\n"
	for i, item := range m.selectList {
		cursor := " "
		if m.cursor == i {
			cursor = cRed.Render(">")
		}

		selection := ""
		if m.selectionMode == modeRemove {
			if _, selected := m.selectedIDs[item.ID]; selected {
				selection = "[x] "
			} else {
				selection = "[ ] "
			}
		}

		line := fmt.Sprintf("%s %s%s %s", cursor, selection, item.Type, item.Text)
		if m.cursor == i {
			ss += cYellow.Render(line) + "\n"
		} else {
			ss += line + "\n"
		}
	}
	
	if m.selectionMode == modeRemove {
		ss += "\n" + cGray.Render("Space to toggle | Enter to remove selected")
	}
	
	return ss
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}