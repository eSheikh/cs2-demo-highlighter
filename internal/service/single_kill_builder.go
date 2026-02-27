package service

import "github.com/eSheikh/cs2-demo-highlighter/internal/model"

type singleKillRule struct {
	highlightType model.HighlightType
	matches       func(kill model.KillEvent) bool
}

var singleKillRules = []singleKillRule{
	{highlightType: model.HighlightKillInSmoke, matches: func(kill model.KillEvent) bool { return kill.IsInSmoke }},
	{highlightType: model.HighlightKillBlinded, matches: func(kill model.KillEvent) bool { return kill.IsBlinded }},
	{highlightType: model.HighlightWallbang, matches: func(kill model.KillEvent) bool { return kill.IsWallbang }},
	{highlightType: model.HighlightNoScope, matches: func(kill model.KillEvent) bool { return kill.IsNoScope }},
	{highlightType: model.HighlightHeadshot, matches: func(kill model.KillEvent) bool { return kill.IsHeadshot }},
}

func (s *HighlightService) buildSingleKillHighlights(demo string, steamID string, kills []model.KillEvent) []model.Highlight {
	items := make([]model.Highlight, 0, len(kills))
	for _, kill := range kills {
		for _, rule := range singleKillRules {
			if !rule.matches(kill) {
				continue
			}
			items = append(items, newSingleKillHighlight(demo, steamID, kill, rule.highlightType))
		}
	}
	return items
}
