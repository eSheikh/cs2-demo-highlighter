package service

import "github.com/eSheikh/cs2-demo-highlighter/internal/model"

func (s *HighlightService) buildHeadshotCollectionHighlights(demo string, steamID string, kills []model.KillEvent) []model.Highlight {
	headshots := filterHeadshots(kills)
	if len(headshots) == 0 {
		return nil
	}

	first := headshots[0]
	last := headshots[len(headshots)-1]

	return []model.Highlight{
		{
			Type:        model.HighlightHeadshotMix,
			Round:       first.Round,
			TickStart:   first.Tick,
			TickEnd:     last.Tick,
			TimeStart:   first.Time.Seconds(),
			TimeEnd:     last.Time.Seconds(),
			Kills:       len(headshots),
			Meta:        map[string]string{"scope": "all_match_headshots"},
			Victims:     collectVictims(headshots),
			Weapon:      "mixed",
			PlayerSlot:  first.KillerSlot,
			SteamID:     steamID,
			Demo:        demo,
			SegmentFrom: first.Tick,
			SegmentTo:   last.Tick,
		},
	}
}

func filterHeadshots(kills []model.KillEvent) []model.KillEvent {
	headshots := make([]model.KillEvent, 0)
	for _, kill := range kills {
		if !kill.IsHeadshot {
			continue
		}
		headshots = append(headshots, kill)
	}
	return headshots
}
