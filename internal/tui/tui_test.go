package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/eSheikh/cs2-demo-highlighter/internal/hlae"
	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

func sampleResult() model.HighlightResult {
	return model.HighlightResult{
		Demo:     "match.dem",
		SteamID:  "7656119",
		TickRate: 64,
		Highlights: []model.Highlight{
			{Type: model.HighlightWallbang, SegmentFrom: 100, SegmentTo: 110, PlayerSlot: 7},
			{Type: model.HighlightWallbang, SegmentFrom: 300, SegmentTo: 310, PlayerSlot: 7},
			{Type: model.HighlightClutchWin, SegmentFrom: 500, SegmentTo: 520, PlayerSlot: 7},
		},
	}
}

func TestCountTypesOrdersAndCounts(t *testing.T) {
	types := countTypes(sampleResult())
	if len(types) != 2 {
		t.Fatalf("expected 2 present types, got %d", len(types))
	}
	if types[0].Type != model.HighlightWallbang || types[0].Count != 2 {
		t.Fatalf("unexpected first type: %+v", types[0])
	}
	if types[1].Type != model.HighlightClutchWin || types[1].Count != 1 {
		t.Fatalf("unexpected second type: %+v", types[1])
	}
	for _, tc := range types {
		if !tc.Enabled {
			t.Fatalf("types should default to enabled: %+v", tc)
		}
	}
}

func TestExtractEventDoneMovesToResults(t *testing.T) {
	m := appModel{state: stateParsing}
	updated, _ := m.Update(extractEvent{done: true, result: sampleResult()})
	got := updated.(appModel)

	if got.state != stateResults {
		t.Fatalf("expected results state, got %d", got.state)
	}
	if len(got.types) != 2 {
		t.Fatalf("expected 2 types, got %d", len(got.types))
	}
	if got.fraction != 1 {
		t.Fatalf("expected fraction 1, got %v", got.fraction)
	}
}

func TestToggleAndSelection(t *testing.T) {
	m := appModel{
		state: stateResults,
		types: countTypes(sampleResult()),
	}

	// disable the first type (wallbang) via space on the cursor.
	updated, _ := m.updateResults(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(appModel)

	sel := m.selection()
	if sel[model.HighlightWallbang] {
		t.Fatalf("wallbang should be disabled after toggle")
	}
	if !sel[model.HighlightClutchWin] {
		t.Fatalf("clutch should remain enabled")
	}
}

func TestGenerateWritesCfg(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m := appModel{
		state:   stateResults,
		options: defaultOptions(),
		result:  sampleResult(),
		types:   countTypes(sampleResult()),
	}

	msg := m.generate(hlae.ModeClips, "clips")
	if msg == "" {
		t.Fatalf("expected a status message")
	}

	data, err := os.ReadFile(filepath.Join(dir, "clips.cfg"))
	if err != nil {
		t.Fatalf("expected clips.cfg written: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("clips.cfg is empty")
	}
}
