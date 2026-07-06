package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
	"github.com/eSheikh/cs2-demo-highlighter/internal/service"
)

type fakeParser struct {
	parsed    model.ParsedDemo
	parseErr  error
	roster    []model.Player
	rosterErr error
	fractions []float64
}

func (f *fakeParser) Parse(_ context.Context, _ string, _ string, onProgress func(float64)) (model.ParsedDemo, error) {
	if onProgress != nil {
		for _, fr := range f.fractions {
			onProgress(fr)
		}
	}
	return f.parsed, f.parseErr
}

func (f *fakeParser) Roster(_ context.Context, _ string) ([]model.Player, error) {
	return f.roster, f.rosterErr
}

func TestExtractBuildsHighlightsAndStreamsProgress(t *testing.T) {
	parser := &fakeParser{
		parsed: model.ParsedDemo{
			Demo:     "match.dem",
			TickRate: 64,
			Kills: []model.KillEvent{
				{Tick: 100, Round: 1, VictimID: "v1", KillerSlot: 7, IsWallbang: true},
			},
		},
		fractions: []float64{0.5, 1},
	}
	eng := New(parser, service.NewHighlightService())

	progress := make(chan Progress, 8)
	result, err := eng.Extract(context.Background(), ExtractOptions{
		DemoPath: "match.dem", SteamID: "steam", Progress: progress,
	})
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(result.Highlights) != 1 || result.Highlights[0].Type != model.HighlightWallbang {
		t.Fatalf("unexpected highlights: %+v", result.Highlights)
	}
	if result.SteamID != "steam" || result.TickRate != 64 {
		t.Fatalf("unexpected metadata: %+v", result)
	}

	got := make([]float64, 0, 2)
	for p := range progress {
		got = append(got, p.Fraction)
	}
	if len(got) == 0 || got[len(got)-1] != 1 {
		t.Fatalf("expected progress ending at 1, got %v", got)
	}
}

func TestExtractPropagatesParseErrorAndClosesProgress(t *testing.T) {
	parser := &fakeParser{parseErr: errors.New("boom")}
	eng := New(parser, service.NewHighlightService())

	progress := make(chan Progress, 1)
	_, err := eng.Extract(context.Background(), ExtractOptions{
		DemoPath: "match.dem", SteamID: "steam", Progress: progress,
	})
	if err == nil {
		t.Fatalf("expected parse error")
	}

	select {
	case _, open := <-progress:
		if open {
			// drain until closed
			for range progress {
			}
		}
	case <-time.After(time.Second):
		t.Fatalf("progress channel was not closed on error")
	}
}

func TestExtractWithoutProgressChannel(t *testing.T) {
	parser := &fakeParser{parsed: model.ParsedDemo{Demo: "m.dem", TickRate: 64}, fractions: []float64{0.5}}
	eng := New(parser, service.NewHighlightService())

	if _, err := eng.Extract(context.Background(), ExtractOptions{DemoPath: "m.dem", SteamID: "steam"}); err != nil {
		t.Fatalf("extract without progress: %v", err)
	}
}

func TestExtractAppliesTypeSelection(t *testing.T) {
	parser := &fakeParser{
		parsed: model.ParsedDemo{
			Demo:     "match.dem",
			TickRate: 64,
			Kills: []model.KillEvent{
				{Tick: 100, Round: 1, VictimID: "v1", KillerSlot: 7, IsWallbang: true},
			},
		},
	}
	eng := New(parser, service.NewHighlightService())

	res, err := eng.Extract(context.Background(), ExtractOptions{
		DemoPath: "match.dem",
		SteamID:  "steam",
		Types:    model.Selection{model.HighlightNoScope: true},
	})
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(res.Highlights) != 0 {
		t.Fatalf("expected wallbang filtered out, got %+v", res.Highlights)
	}
}

func TestRosterDelegates(t *testing.T) {
	want := []model.Player{{SteamID: "1", Name: "a"}}
	eng := New(&fakeParser{roster: want}, service.NewHighlightService())

	got, err := eng.Roster(context.Background(), "m.dem")
	if err != nil {
		t.Fatalf("roster: %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("unexpected roster: %+v", got)
	}
}
