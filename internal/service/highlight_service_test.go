package service

import (
	"testing"
	"time"

	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

func TestBuildHighlightsBuildsSingleAndMultiKill(t *testing.T) {
	svc := NewHighlightService()
	kills := []model.KillEvent{
		{
			Tick:       100,
			Time:       10 * time.Second,
			Round:      1,
			VictimID:   "v1",
			Weapon:     "ak47",
			KillerSlot: 7,
			IsWallbang: true,
		},
		{
			Tick:       120,
			Time:       12 * time.Second,
			Round:      1,
			VictimID:   "v2",
			Weapon:     "ak47",
			KillerSlot: 7,
			IsNoScope:  true,
		},
	}

	result := svc.BuildHighlights("match.dem", "7656119", 64, kills)

	if result.Demo != "match.dem" || result.SteamID != "7656119" || result.TickRate != 64 {
		t.Fatalf("unexpected metadata: %+v", result)
	}
	if len(result.Highlights) != 3 {
		t.Fatalf("expected 3 highlights, got %d", len(result.Highlights))
	}
	if result.Highlights[0].Type != model.HighlightWallbang {
		t.Fatalf("expected first single highlight type %s, got %s", model.HighlightWallbang, result.Highlights[0].Type)
	}
	if result.Highlights[1].Type != model.HighlightNoScope {
		t.Fatalf("expected second single highlight type %s, got %s", model.HighlightNoScope, result.Highlights[1].Type)
	}
	if result.Highlights[2].Type != model.HighlightMultiKill {
		t.Fatalf("expected multikill highlight, got %s", result.Highlights[2].Type)
	}
}

func TestGroupKillsByRoundSplitsOnlyByRound(t *testing.T) {
	kills := []model.KillEvent{
		{Tick: 100, Time: 1 * time.Second, Round: 1},
		{Tick: 110, Time: 4 * time.Second, Round: 1},
		{Tick: 200, Time: 20 * time.Second, Round: 1},
		{Tick: 300, Time: 21 * time.Second, Round: 2},
	}

	groups := groupKillsByRound(kills)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if len(groups[0]) != 3 {
		t.Fatalf("expected first group size 3, got %d", len(groups[0]))
	}
	if groups[1][0].Tick != 300 {
		t.Fatalf("expected second group to start at tick 300, got %d", groups[1][0].Tick)
	}
	if groups[1][0].Round != 2 {
		t.Fatalf("expected second group to be round 2, got round %d", groups[1][0].Round)
	}
}

func TestBuildClutchWinHighlights(t *testing.T) {
	svc := NewHighlightService()
	kills := []model.KillEvent{
		{
			Tick:               500,
			Time:               30 * time.Second,
			Round:              7,
			VictimID:           "v1",
			Weapon:             "ak47",
			KillerSlot:         9,
			AlliesAliveBefore:  2,
			EnemiesAliveBefore: 4,
			RoundWon:           true,
		},
		{
			Tick:               540,
			Time:               31 * time.Second,
			Round:              7,
			VictimID:           "v2",
			Weapon:             "ak47",
			KillerSlot:         9,
			AlliesAliveBefore:  1,
			EnemiesAliveBefore: 3,
			RoundWon:           true,
		},
		{
			Tick:               580,
			Time:               32 * time.Second,
			Round:              7,
			VictimID:           "v3",
			Weapon:             "ak47",
			KillerSlot:         9,
			AlliesAliveBefore:  1,
			EnemiesAliveBefore: 2,
			RoundWon:           true,
		},
		{
			Tick:               620,
			Time:               33 * time.Second,
			Round:              7,
			VictimID:           "v4",
			Weapon:             "ak47",
			KillerSlot:         9,
			AlliesAliveBefore:  1,
			EnemiesAliveBefore: 1,
			RoundWon:           true,
		},
	}

	highlights := svc.buildClutchWinHighlights("match.dem", "steam", kills)
	if len(highlights) != 1 {
		t.Fatalf("expected 1 clutch highlight, got %d", len(highlights))
	}

	highlight := highlights[0]
	if highlight.Type != model.HighlightClutchWin {
		t.Fatalf("expected clutch highlight type, got %s", highlight.Type)
	}
	if highlight.SegmentFrom != 540 || highlight.SegmentTo != 620 {
		t.Fatalf("unexpected clutch segment range: %d..%d", highlight.SegmentFrom, highlight.SegmentTo)
	}
	if highlight.Meta["clutch"] != "1v3" {
		t.Fatalf("expected clutch meta 1v3, got %q", highlight.Meta["clutch"])
	}
	if highlight.Kills != 3 {
		t.Fatalf("expected 3 kills in clutch sequence, got %d", highlight.Kills)
	}
}

func TestBuildHeadshotCollectionHighlights(t *testing.T) {
	svc := NewHighlightService()
	kills := []model.KillEvent{
		{Tick: 100, Time: 10 * time.Second, Round: 1, VictimID: "v1", KillerSlot: 7, IsHeadshot: true},
		{Tick: 200, Time: 20 * time.Second, Round: 2, VictimID: "v2", KillerSlot: 7, IsHeadshot: false},
		{Tick: 300, Time: 30 * time.Second, Round: 3, VictimID: "v3", KillerSlot: 7, IsHeadshot: true},
	}

	highlights := svc.buildHeadshotCollectionHighlights("match.dem", "steam", kills)
	if len(highlights) != 1 {
		t.Fatalf("expected 1 headshot collection highlight, got %d", len(highlights))
	}

	highlight := highlights[0]
	if highlight.Type != model.HighlightHeadshotMix {
		t.Fatalf("expected headshot collection highlight type, got %s", highlight.Type)
	}
	if highlight.SegmentFrom != 100 || highlight.SegmentTo != 300 {
		t.Fatalf("unexpected headshot segment range: %d..%d", highlight.SegmentFrom, highlight.SegmentTo)
	}
	if highlight.Kills != 2 {
		t.Fatalf("expected 2 headshots in collection, got %d", highlight.Kills)
	}
	if highlight.Meta["scope"] != "all_match_headshots" {
		t.Fatalf("unexpected headshot collection meta: %q", highlight.Meta["scope"])
	}
}
