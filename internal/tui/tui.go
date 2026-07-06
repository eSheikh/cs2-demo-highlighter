// Package tui is an interactive front-end over the engine core: pick a demo,
// pick a player, watch parsing progress, choose highlight types, generate cfg.
package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eSheikh/cs2-demo-highlighter/internal/engine"
	"github.com/eSheikh/cs2-demo-highlighter/internal/hlae"
	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

type state int

const (
	statePicker state = iota
	stateRosterLoading
	stateRoster
	stateParsing
	stateResults
)

const (
	focusTypes = iota
	focusOutput
)

// Run starts the interactive program. demoArg, if a path, sets the file
// picker's starting directory.
func Run(eng *engine.Engine, demoArg string) error {
	m := newModel(eng, demoArg)
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func newModel(eng *engine.Engine, demoArg string) appModel {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".dem"}
	fp.CurrentDirectory = startDir(demoArg)
	fp.AutoHeight = false

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	out := textinput.New()
	out.SetValue("clips")
	out.CharLimit = 64

	players := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	players.Title = "Players"
	players.SetShowHelp(false)

	m := appModel{
		eng:      eng,
		state:    statePicker,
		options:  defaultOptions(),
		picker:   fp,
		spin:     sp,
		players:  players,
		progress: progress.New(progress.WithDefaultGradient()),
		output:   out,
		mode:     hlae.ModeClips,
	}
	// A demo file argument skips the picker and loads its roster directly.
	if isDemoFile(demoArg) {
		m.demoPath = demoArg
		m.state = stateRosterLoading
	}
	return m
}

func isDemoFile(path string) bool {
	if filepath.Ext(path) != ".dem" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func startDir(demoArg string) string {
	if demoArg != "" {
		if info, err := os.Stat(demoArg); err == nil && info.IsDir() {
			return demoArg
		}
		if dir := filepath.Dir(demoArg); dir != "" {
			return dir
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
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

type playerItem struct{ player model.Player }

func (i playerItem) Title() string       { return i.player.Name }
func (i playerItem) Description() string  { return i.player.SteamID }
func (i playerItem) FilterValue() string { return i.player.Name }

type appModel struct {
	eng     *engine.Engine
	state   state
	options hlae.Options
	width   int
	height  int

	picker   filepicker.Model
	spin     spinner.Model
	players  list.Model
	progress progress.Model
	output   textinput.Model

	demoPath string
	err      error

	fraction float64
	events   chan extractEvent

	result       model.HighlightResult
	types        []typeCount
	typeCursor   int
	mode         hlae.Mode
	resultsFocus int
	generatedMsg string
}

func (m appModel) Init() tea.Cmd {
	if m.state == stateRosterLoading {
		return tea.Batch(m.spin.Tick, rosterCmd(m.eng, m.demoPath))
	}
	return tea.Batch(m.picker.Init(), m.spin.Tick)
}

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
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.applySizes()
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	case rosterMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = statePicker
			return m, nil
		}
		if len(msg.players) == 0 {
			m.err = fmt.Errorf("no players found in demo")
			m.state = statePicker
			return m, nil
		}
		m.err = nil
		items := make([]list.Item, len(msg.players))
		for i, p := range msg.players {
			items[i] = playerItem{player: p}
		}
		m.players.SetItems(items)
		m.state = stateRoster
		return m, nil
	case extractEvent:
		if msg.err != nil {
			m.err = msg.err
			m.state = statePicker
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

	switch m.state {
	case statePicker:
		return m.updatePicker(msg)
	case stateRoster:
		return m.updateRoster(msg)
	case stateResults:
		if key, ok := msg.(tea.KeyMsg); ok {
			return m.updateResults(key)
		}
	}
	return m, nil
}

func (m *appModel) applySizes() {
	m.picker.SetHeight(max(m.height-8, 3))
	m.players.SetSize(max(m.width-6, 10), max(m.height-8, 3))
	m.progress.Width = min(max(m.width-12, 20), 60)
	m.output.Width = 24
}

func (m appModel) updatePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)
	if ok, path := m.picker.DidSelectFile(msg); ok {
		m.demoPath = path
		m.err = nil
		m.state = stateRosterLoading
		return m, tea.Batch(cmd, m.spin.Tick, rosterCmd(m.eng, path))
	}
	return m, cmd
}

func (m appModel) updateRoster(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && m.players.FilterState() != list.Filtering {
		switch key.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			if item, ok := m.players.SelectedItem().(playerItem); ok {
				m.events = make(chan extractEvent, 32)
				m.fraction = 0
				m.err = nil
				m.state = stateParsing
				return m, tea.Batch(
					m.spin.Tick,
					startExtract(m.eng, m.demoPath, item.player.SteamID, m.events),
					waitExtract(m.events),
				)
			}
		}
	}
	var cmd tea.Cmd
	m.players, cmd = m.players.Update(msg)
	return m, cmd
}

func (m appModel) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.resultsFocus == focusOutput {
		switch msg.String() {
		case "tab", "esc":
			m.resultsFocus = focusTypes
			m.output.Blur()
		case "enter":
			m.generatedMsg = m.generate(m.mode, m.output.Value())
		default:
			var cmd tea.Cmd
			m.output, cmd = m.output.Update(msg)
			return m, cmd
		}
		return m, nil
	}

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
	case "m":
		if m.mode == hlae.ModeClips {
			m.mode = hlae.ModeMontage
		} else {
			m.mode = hlae.ModeClips
		}
	case "tab":
		m.resultsFocus = focusOutput
		return m, m.output.Focus()
	case "enter", "g":
		m.generatedMsg = m.generate(m.mode, m.output.Value())
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
	if name == "" {
		name = modeName(mode)
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
	return okStyle.Render(fmt.Sprintf("saved %s", path))
}

func modeName(mode hlae.Mode) string {
	if mode == hlae.ModeMontage {
		return "montage"
	}
	return "clips"
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
	case statePicker:
		return chrome(m.width, m.height, "1/4  Demo", m.bodyPicker(),
			"↑/↓ navigate   enter select   ctrl+c quit")
	case stateRosterLoading:
		return chrome(m.width, m.height, "2/4  Player", m.spin.View()+" reading roster…",
			"ctrl+c quit")
	case stateRoster:
		return chrome(m.width, m.height, "2/4  Player", m.players.View(),
			"↑/↓ move   / filter   enter parse   ctrl+c quit")
	case stateParsing:
		return chrome(m.width, m.height, "3/4  Parsing", m.bodyParsing(),
			"parsing…   ctrl+c quit")
	case stateResults:
		return chrome(m.width, m.height, "4/4  Output", m.bodyResults(), m.resultsFooter())
	}
	return ""
}

func (m appModel) bodyPicker() string {
	s := titleStyle.Render("Select a demo (.dem)") + "\n"
	s += dimStyle.Render(m.picker.CurrentDirectory) + "\n\n"
	s += m.picker.View()
	if m.err != nil {
		s += "\n\n" + errStyle.Render("error: "+m.err.Error())
	}
	return s
}

func (m appModel) bodyParsing() string {
	s := titleStyle.Render("Parsing demo") + "\n"
	s += dimStyle.Render(filepath.Base(m.demoPath)) + "\n\n"
	s += m.progress.ViewAs(m.fraction) + "\n\n"
	s += m.spin.View() + dimStyle.Render(fmt.Sprintf(" %.0f%%", m.fraction*100))
	return s
}

func (m appModel) bodyResults() string {
	s := titleStyle.Render("Highlights") + "  "
	s += dimStyle.Render(fmt.Sprintf("%s — %d total", m.result.Demo, len(m.result.Highlights))) + "\n\n"

	if len(m.types) == 0 {
		s += dimStyle.Render("no highlights found") + "\n"
	}
	for i, t := range m.types {
		check := "[ ]"
		if t.Enabled {
			check = "[x]"
		}
		cursor := "  "
		line := fmt.Sprintf("%s %-16s %s", check, t.Type, dimStyle.Render(fmt.Sprintf("×%d", t.Count)))
		if i == m.typeCursor && m.resultsFocus == focusTypes {
			cursor = cursorStyle.Render("> ")
			line = selectedStyle.Render(line)
		}
		s += cursor + line + "\n"
	}

	clips, montage := inactiveTab.Render("clips"), inactiveTab.Render("montage")
	if m.mode == hlae.ModeMontage {
		montage = activeTab.Render("montage")
	} else {
		clips = activeTab.Render("clips")
	}
	s += "\n" + "Mode  " + clips + " " + montage + dimStyle.Render("   (m to toggle)") + "\n"
	s += "Output  " + m.output.View() + dimStyle.Render(".cfg")
	if m.generatedMsg != "" {
		s += "\n\n" + m.generatedMsg
	}
	return s
}

func (m appModel) resultsFooter() string {
	if m.resultsFocus == focusOutput {
		return "type name   enter generate   tab/esc back   ctrl+c quit"
	}
	return "↑/↓ move   space toggle   m mode   tab name   enter generate   q quit"
}
