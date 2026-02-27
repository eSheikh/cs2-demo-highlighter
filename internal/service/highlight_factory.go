package service

import "github.com/eSheikh/cs2-demo-highlighter/internal/model"

func newSingleKillHighlight(demo string, steamID string, kill model.KillEvent, highlightType model.HighlightType) model.Highlight {
	return model.Highlight{
		Type:        highlightType,
		Round:       kill.Round,
		TickStart:   kill.Tick,
		TickEnd:     kill.Tick,
		TimeStart:   kill.Time.Seconds(),
		TimeEnd:     kill.Time.Seconds(),
		Victims:     []string{kill.VictimID},
		Weapon:      kill.Weapon,
		PlayerSlot:  kill.KillerSlot,
		SteamID:     steamID,
		Demo:        demo,
		SegmentFrom: kill.Tick,
		SegmentTo:   kill.Tick,
	}
}

func newMultiKillHighlight(demo string, steamID string, kills []model.KillEvent) model.Highlight {
	first := kills[0]
	last := kills[len(kills)-1]

	return model.Highlight{
		Type:        model.HighlightMultiKill,
		Round:       first.Round,
		TickStart:   first.Tick,
		TickEnd:     last.Tick,
		TimeStart:   first.Time.Seconds(),
		TimeEnd:     last.Time.Seconds(),
		Kills:       len(kills),
		KillTicks:   collectKillTicks(kills),
		Victims:     collectVictims(kills),
		Weapon:      last.Weapon,
		PlayerSlot:  first.KillerSlot,
		SteamID:     steamID,
		Demo:        demo,
		SegmentFrom: first.Tick,
		SegmentTo:   last.Tick,
	}
}

func collectVictims(kills []model.KillEvent) []string {
	victims := make([]string, 0, len(kills))
	for _, kill := range kills {
		victims = append(victims, kill.VictimID)
	}
	return victims
}

func collectKillTicks(kills []model.KillEvent) []int {
	ticks := make([]int, 0, len(kills))
	for _, kill := range kills {
		ticks = append(ticks, kill.Tick)
	}
	return ticks
}
