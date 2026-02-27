package service

import (
	"context"

	"cs2-demo-highlighter/internal/model"
)

type DemoParser interface {
	Parse(ctx context.Context, demoPath string, steamID string) (model.ParsedDemo, error)
}

type HighlightBuilder interface {
	BuildHighlights(demo string, steamID string, tickRate float64, kills []model.KillEvent) model.HighlightResult
}

type Orchestrator struct {
	parser  DemoParser
	builder HighlightBuilder
}

func NewOrchestrator(parser DemoParser, builder HighlightBuilder) *Orchestrator {
	return &Orchestrator{parser: parser, builder: builder}
}

func (o *Orchestrator) Build(ctx context.Context, demoPath string, steamID string) (model.HighlightResult, error) {
	result, err := o.parser.Parse(ctx, demoPath, steamID)
	if err != nil {
		return model.HighlightResult{}, err
	}
	return o.builder.BuildHighlights(result.Demo, steamID, result.TickRate, result.Kills), nil
}
