package service

import (
	"slices"

	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

func (s *HighlightService) buildMultiKillHighlights(demo string, steamID string, kills []model.KillEvent) []model.Highlight {
	groups := groupKillsByRound(kills)
	items := make([]model.Highlight, 0, len(groups))
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		items = append(items, newMultiKillHighlight(demo, steamID, group))
	}
	return items
}

func groupKillsByRound(kills []model.KillEvent) [][]model.KillEvent {
	groups := make([][]model.KillEvent, 0)
	roundKills := make([]model.KillEvent, 0)

	flush := func() {
		if len(roundKills) == 0 {
			return
		}
		groups = append(groups, slices.Clone(roundKills))
		roundKills = roundKills[:0]
	}

	for _, kill := range kills {
		if len(roundKills) > 0 && roundKills[len(roundKills)-1].Round != kill.Round {
			flush()
		}
		roundKills = append(roundKills, kill)
	}
	flush()

	return groups
}
