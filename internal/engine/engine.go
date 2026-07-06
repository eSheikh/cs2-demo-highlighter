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
	BuildHighlights(demo string, steamID string, tickRate float64, kills []model.KillEvent, selection model.Selection) model.HighlightResult
}

// Progress reports parsing advancement as a 0..1 fraction.
type Progress struct {
	Fraction float64
}

// ExtractOptions configures a single extraction run. Types selects which
// highlight types to keep (empty = all). Progress, if non-nil, receives updates
// as parsing advances and is closed on return.
type ExtractOptions struct {
	DemoPath string
	SteamID  string
	Types    model.Selection
	Progress chan<- Progress
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

// Extract parses the demo and builds the selected highlights for opts.SteamID.
// If opts.Progress is non-nil it receives updates as parsing advances and is
// closed on return; the send is non-blocking, so a slow consumer only drops
// intermediate updates.
func (e *Engine) Extract(ctx context.Context, opts ExtractOptions) (model.HighlightResult, error) {
	if opts.Progress != nil {
		defer close(opts.Progress)
	}

	onProgress := func(fraction float64) {
		if opts.Progress == nil {
			return
		}
		select {
		case opts.Progress <- Progress{Fraction: fraction}:
		default:
		}
	}

	parsed, err := e.parser.Parse(ctx, opts.DemoPath, opts.SteamID, onProgress)
	if err != nil {
		return model.HighlightResult{}, err
	}

	return e.builder.BuildHighlights(parsed.Demo, opts.SteamID, parsed.TickRate, parsed.Kills, opts.Types), nil
}
