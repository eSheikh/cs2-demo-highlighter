package service

import (
	"fmt"
	"sort"

	"cs2-demo-highlighter/internal/model"
)

type roundKillWindow struct {
	kills              []model.KillEvent
	firstClutchKillIdx int
	maxEnemiesInClutch int
	wonRound           bool
}

func (s *HighlightService) buildClutchWinHighlights(demo string, steamID string, kills []model.KillEvent) []model.Highlight {
	rounds := collectRoundsForClutch(kills)
	if len(rounds) == 0 {
		return nil
	}

	roundNumbers := make([]int, 0, len(rounds))
	for round := range rounds {
		roundNumbers = append(roundNumbers, round)
	}
	sort.Ints(roundNumbers)

	items := make([]model.Highlight, 0, len(roundNumbers))
	for _, round := range roundNumbers {
		ctx := rounds[round]
		clutchKills := resolveClutchKills(ctx)
		if len(clutchKills) == 0 || !ctx.wonRound {
			continue
		}
		items = append(items, newClutchWinHighlight(demo, steamID, ctx.maxEnemiesInClutch, clutchKills))
	}

	return items
}

func collectRoundsForClutch(kills []model.KillEvent) map[int]*roundKillWindow {
	rounds := make(map[int]*roundKillWindow)

	for _, kill := range kills {
		ctx := roundWindowFor(rounds, kill.Round)

		ctx.kills = append(ctx.kills, kill)
		ctx.wonRound = ctx.wonRound || kill.RoundWon
		if !isClutchStartKill(kill) {
			continue
		}
		if ctx.firstClutchKillIdx < 0 {
			ctx.firstClutchKillIdx = len(ctx.kills) - 1
		}
		ctx.maxEnemiesInClutch = max(ctx.maxEnemiesInClutch, kill.EnemiesAliveBefore)
	}

	return rounds
}

func roundWindowFor(rounds map[int]*roundKillWindow, round int) *roundKillWindow {
	if existing := rounds[round]; existing != nil {
		return existing
	}

	window := &roundKillWindow{
		kills:              make([]model.KillEvent, 0),
		firstClutchKillIdx: -1,
	}
	rounds[round] = window
	return window
}

func resolveClutchKills(ctx *roundKillWindow) []model.KillEvent {
	if ctx == nil || ctx.firstClutchKillIdx < 0 || ctx.firstClutchKillIdx >= len(ctx.kills) {
		return nil
	}
	return ctx.kills[ctx.firstClutchKillIdx:]
}

func isClutchStartKill(kill model.KillEvent) bool {
	return kill.AlliesAliveBefore == 1 && kill.EnemiesAliveBefore >= 2
}

func newClutchWinHighlight(demo string, steamID string, maxEnemies int, clutchKills []model.KillEvent) model.Highlight {
	first := clutchKills[0]
	last := clutchKills[len(clutchKills)-1]

	return model.Highlight{
		Type:      model.HighlightClutchWin,
		Round:     first.Round,
		TickStart: first.Tick,
		TickEnd:   last.Tick,
		TimeStart: first.Time.Seconds(),
		TimeEnd:   last.Time.Seconds(),
		Kills:     len(clutchKills),
		Meta: map[string]string{
			"clutch": fmt.Sprintf("1v%d", maxEnemies),
		},
		Victims:     collectVictims(clutchKills),
		Weapon:      last.Weapon,
		PlayerSlot:  first.KillerSlot,
		SteamID:     steamID,
		Demo:        demo,
		SegmentFrom: first.Tick,
		SegmentTo:   last.Tick,
	}
}
