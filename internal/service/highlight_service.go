package service

import (
	"slices"

	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

type HighlightService struct{}

func NewHighlightService() *HighlightService {
	return &HighlightService{}
}

func (s *HighlightService) BuildHighlights(demo string, steamID string, tickRate float64, kills []model.KillEvent, selection model.Selection) model.HighlightResult {
	highlights := slices.Concat(
		s.buildSingleKillHighlights(demo, steamID, kills),
		s.buildMultiKillHighlights(demo, steamID, kills),
		s.buildClutchWinHighlights(demo, steamID, kills),
		s.buildHeadshotCollectionHighlights(demo, steamID, kills),
	)

	return model.HighlightResult{
		Demo:       demo,
		SteamID:    steamID,
		TickRate:   tickRate,
		Highlights: filterBySelection(highlights, selection),
	}
}

func filterBySelection(highlights []model.Highlight, selection model.Selection) []model.Highlight {
	if len(selection) == 0 {
		return highlights
	}
	filtered := make([]model.Highlight, 0, len(highlights))
	for _, h := range highlights {
		if selection.Enabled(h.Type) {
			filtered = append(filtered, h)
		}
	}
	return filtered
}
