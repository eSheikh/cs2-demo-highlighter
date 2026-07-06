// Package tui is an interactive front-end over the engine core: pick a demo,
// pick a player, watch parsing progress, choose highlight types, generate cfg.
package tui

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/eSheikh/cs2-demo-highlighter/internal/engine"
	"github.com/eSheikh/cs2-demo-highlighter/internal/hlae"
	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

type state int

const (
	stateDemo state = iota
	stateRoster
	stateParsing
	stateResults
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	okStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
)

// Run starts the interactive program, defaulting the demo path field to demoArg.
func Run(eng *engine.Engine, demoArg string) error {
	input := textinput.New()
	input.Placeholder = "/path/to/match.dem"
	input.SetValue(demoArg)
	input.Focus()
	input.CursorEnd()

	m := appModel{
		eng:      eng,
		state:    stateDemo,
		input:    input,
		progress: progress.New(progress.WithDefaultGradient()),
		options:  defaultOptions(),
	}

	_, err := tea.NewProgram(m).Run()
	return err
}

func defaultOptions() hlae.Options {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	return hlae.Options{
		FrameRate:       60,
		OutputPath:      cwd,
		FFmpegPreset:    "afxFfmpegYuv420p",
		PreRollSeconds:  3,
		PostRollSeconds: 2,
		KillGapSeconds:  10,
	}
}

type typeCount struct {
	Type    model.HighlightType
	Count   int
	Enabled bool
}

type appModel struct {
	eng     *engine.Engine
	state   state
	options hlae.Options

	input   textinput.Model
	err     error
	loading bool

	roster       []model.Player
	rosterCursor int

	progress progress.Model
	fraction float64
	events   chan extractEvent

	result       model.HighlightResult
	types        []typeCount
	typeCursor   int
	generatedMsg string
}

func (m appModel) Init() tea.Cmd { return textinput.Blink }

// --- messages / commands ---

type rosterMsg struct {
	players []model.Player
	err     error
}

type extractEvent struct {
	fraction float64
	done     bool
	result   model.HighlightResult
	err      error
}

func rosterCmd(eng *engine.Engine, demoPath string) tea.Cmd {
	return func() tea.Msg {
		players, err := eng.Roster(context.Background(), demoPath)
		return rosterMsg{players: players, err: err}
	}
}

func startExtract(eng *engine.Engine, demoPath, steamID string, events chan extractEvent) tea.Cmd {
	return func() tea.Msg {
		progressCh := make(chan engine.Progress, 16)
		go func() {
			for p := range progressCh {
				events <- extractEvent{fraction: p.Fraction}
			}
		}()
		go func() {
			result, err := eng.Extract(context.Background(), engine.ExtractOptions{
				DemoPath: demoPath,
				SteamID:  steamID,
				Progress: progressCh,
			})
			events <- extractEvent{done: true, result: result, err: err}
		}()
		return nil
	}
}

func waitExtract(events chan extractEvent) tea.Cmd {
	return func() tea.Msg {
		return <-events
	}
}

// --- update ---

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		switch m.state {
		case stateDemo:
			return m.updateDemo(msg)
		case stateRoster:
			return m.updateRoster(msg)
		case stateResults:
			return m.updateResults(msg)
		}
	case rosterMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if len(msg.players) == 0 {
			m.err = fmt.Errorf("no players found in demo")
			return m, nil
		}
		m.err = nil
		m.roster = msg.players
		m.rosterCursor = 0
		m.state = stateRoster
		return m, nil
	case extractEvent:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateDemo
			return m, nil
		}
		if msg.done {
			m.result = msg.result
			m.types = countTypes(msg.result)
			m.typeCursor = 0
			m.fraction = 1
			m.state = stateResults
			return m, nil
		}
		m.fraction = msg.fraction
		return m, waitExtract(m.events)
	}
	return m, nil
}

func (m appModel) updateDemo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		demo := m.input.Value()
		if demo == "" {
			return m, nil
		}
		m.loading = true
		m.err = nil
		return m, rosterCmd(m.eng, demo)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m appModel) updateRoster(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.rosterCursor > 0 {
			m.rosterCursor--
		}
	case "down", "j":
		if m.rosterCursor < len(m.roster)-1 {
			m.rosterCursor++
		}
	case "enter":
		m.events = make(chan extractEvent, 32)
		m.fraction = 0
		m.err = nil
		m.state = stateParsing
		player := m.roster[m.rosterCursor]
		return m, tea.Batch(
			startExtract(m.eng, m.input.Value(), player.SteamID, m.events),
			waitExtract(m.events),
		)
	}
	return m, nil
}

func (m appModel) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.typeCursor > 0 {
			m.typeCursor--
		}
	case "down", "j":
		if m.typeCursor < len(m.types)-1 {
			m.typeCursor++
		}
	case " ":
		if len(m.types) > 0 {
			m.types[m.typeCursor].Enabled = !m.types[m.typeCursor].Enabled
		}
	case "c":
		m.generatedMsg = m.generate(hlae.ModeClips, "clips")
	case "m":
		m.generatedMsg = m.generate(hlae.ModeMontage, "montage")
	}
	return m, nil
}

func (m appModel) selection() model.Selection {
	selection := make(model.Selection)
	for _, t := range m.types {
		if t.Enabled {
			selection[t.Type] = true
		}
	}
	return selection
}

func (m appModel) generate(mode hlae.Mode, name string) string {
	selection := m.selection()
	if len(selection) == 0 {
		return errStyle.Render("select at least one highlight type")
	}
	path := name + ".cfg"
	content := hlae.BuildTarget(m.result, m.options, hlae.Target{
		Mode:  mode,
		Types: selection,
		Path:  path,
		Name:  name,
	})
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return errStyle.Render("write failed: " + err.Error())
	}
	return okStyle.Render(fmt.Sprintf("saved %s (%s)", path, name))
}

func countTypes(result model.HighlightResult) []typeCount {
	counts := make(map[model.HighlightType]int)
	for _, h := range result.Highlights {
		counts[h.Type]++
	}
	types := make([]typeCount, 0, len(counts))
	for _, t := range model.AllHighlightTypes() {
		if counts[t] > 0 {
			types = append(types, typeCount{Type: t, Count: counts[t], Enabled: true})
		}
	}
	return types
}

// --- view ---

func (m appModel) View() string {
	switch m.state {
	case stateDemo:
		return m.viewDemo()
	case stateRoster:
		return m.viewRoster()
	case stateParsing:
		return m.viewParsing()
	case stateResults:
		return m.viewResults()
	}
	return ""
}

func (m appModel) viewDemo() string {
	s := titleStyle.Render("CS2 Demo Highlighter") + "\n\n"
	s += "Demo path:\n" + m.input.View() + "\n\n"
	if m.loading {
		s += dimStyle.Render("reading roster…") + "\n"
	}
	if m.err != nil {
		s += errStyle.Render("error: "+m.err.Error()) + "\n"
	}
	s += dimStyle.Render("enter: continue   ctrl+c: quit")
	return s + "\n"
}

func (m appModel) viewRoster() string {
	s := titleStyle.Render("Select player") + "\n\n"
	for i, p := range m.roster {
		cursor := "  "
		line := fmt.Sprintf("%s  %s", p.Name, dimStyle.Render(p.SteamID))
		if i == m.rosterCursor {
			cursor = cursorStyle.Render("> ")
			line = selectedStyle.Render(p.Name) + "  " + dimStyle.Render(p.SteamID)
		}
		s += cursor + line + "\n"
	}
	s += "\n" + dimStyle.Render("↑/↓: move   enter: parse   q: quit")
	return s + "\n"
}

func (m appModel) viewParsing() string {
	s := titleStyle.Render("Parsing demo") + "\n\n"
	s += m.progress.ViewAs(m.fraction) + "\n\n"
	s += dimStyle.Render(fmt.Sprintf("%.0f%%", m.fraction*100))
	return s + "\n"
}

func (m appModel) viewResults() string {
	s := titleStyle.Render("Highlights") + "\n"
	s += dimStyle.Render(fmt.Sprintf("%s — %d highlights", m.result.Demo, len(m.result.Highlights))) + "\n\n"

	if len(m.types) == 0 {
		s += dimStyle.Render("no highlights found") + "\n"
	}
	for i, t := range m.types {
		check := "[ ]"
		if t.Enabled {
			check = "[x]"
		}
		cursor := "  "
		label := fmt.Sprintf("%s %-16s %s", check, t.Type, dimStyle.Render(fmt.Sprintf("×%d", t.Count)))
		if i == m.typeCursor {
			cursor = cursorStyle.Render("> ")
			label = selectedStyle.Render(label)
		}
		s += cursor + label + "\n"
	}

	if m.generatedMsg != "" {
		s += "\n" + m.generatedMsg + "\n"
	}
	s += "\n" + dimStyle.Render("↑/↓: move   space: toggle   c: clips   m: montage   q: quit")
	return s + "\n"
}
