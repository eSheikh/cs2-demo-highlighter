// Package engine is the core highlight-extraction API: it turns a demo file
// into highlight metadata. It performs no file I/O — callers (CLI, TUI, worker)
// decide what to do with the result.
package engine

import (
	"context"

	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

type DemoParser interface {
	Parse(ctx context.Context, demoPath string, steamID string, onProgress func(float64)) (model.ParsedDemo, error)
	Roster(ctx context.Context, demoPath string) ([]model.Player, error)
}

type HighlightBuilder interface {
	BuildHighlights(demo string, steamID string, tickRate float64, kills []model.KillEvent) model.HighlightResult
}

// Progress reports parsing advancement as a 0..1 fraction.
type Progress struct {
	Fraction float64
}

type Engine struct {
	parser  DemoParser
	builder HighlightBuilder
}

func New(parser DemoParser, builder HighlightBuilder) *Engine {
	return &Engine{parser: parser, builder: builder}
}

// Roster lists the players in the demo so a caller can pick one before extracting.
func (e *Engine) Roster(ctx context.Context, demoPath string) ([]model.Player, error) {
	return e.parser.Roster(ctx, demoPath)
}

// Extract parses the demo and builds highlights for steamID. If progress is
// non-nil it receives updates as parsing advances and is closed on return; the
// send is non-blocking, so a slow consumer only drops intermediate updates.
func (e *Engine) Extract(ctx context.Context, demoPath string, steamID string, progress chan<- Progress) (model.HighlightResult, error) {
	if progress != nil {
		defer close(progress)
	}

	onProgress := func(fraction float64) {
		if progress == nil {
			return
		}
		select {
		case progress <- Progress{Fraction: fraction}:
		default:
		}
	}

	parsed, err := e.parser.Parse(ctx, demoPath, steamID, onProgress)
	if err != nil {
		return model.HighlightResult{}, err
	}

	return e.builder.BuildHighlights(parsed.Demo, steamID, parsed.TickRate, parsed.Kills), nil
}
