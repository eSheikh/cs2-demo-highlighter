package service

import (
	"slices"

	"cs2-demo-highlighter/internal/model"
)

type HighlightService struct{}

func NewHighlightService() *HighlightService {
	return &HighlightService{}
}

func (s *HighlightService) BuildHighlights(demo string, steamID string, tickRate float64, kills []model.KillEvent) model.HighlightResult {
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
		Highlights: highlights,
	}
}
