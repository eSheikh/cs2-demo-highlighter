package model

import "time"

type HighlightType string

const (
	HighlightKillInSmoke HighlightType = "kill_in_smoke"
	HighlightMultiKill   HighlightType = "round_multikill"
	HighlightKillBlinded HighlightType = "kill_blinded"
	HighlightWallbang    HighlightType = "wallbang"
	HighlightNoScope     HighlightType = "noscope"
	HighlightHeadshot    HighlightType = "headshot_kill"
	HighlightClutchWin   HighlightType = "clutch_win"
)

func AllHighlightTypes() []HighlightType {
	return []HighlightType{
		HighlightKillInSmoke,
		HighlightKillBlinded,
		HighlightWallbang,
		HighlightNoScope,
		HighlightHeadshot,
		HighlightMultiKill,
		HighlightClutchWin,
	}
}

// Selection is the set of highlight types to keep. An empty selection means
// "all types enabled", so the zero value is a no-op filter.
type Selection map[HighlightType]bool

func (s Selection) Enabled(t HighlightType) bool {
	if len(s) == 0 {
		return true
	}
	return s[t]
}

type KillEvent struct {
	Tick       int
	Time       time.Duration
	Round      int
	KillerID   string
	KillerSlot int
	VictimID   string
	Weapon     string
	IsInSmoke  bool
	IsBlinded  bool
	IsWallbang bool
	IsNoScope  bool
	IsHeadshot bool
	KillerTeam int
	RoundWon   bool

	AlliesAliveBefore  int
	EnemiesAliveBefore int
}

type Highlight struct {
	Type        HighlightType     `json:"type"`
	Round       int               `json:"round"`
	TickStart   int               `json:"tick_start"`
	TickEnd     int               `json:"tick_end"`
	TimeStart   float64           `json:"time_start_sec"`
	TimeEnd     float64           `json:"time_end_sec"`
	Kills       int               `json:"kills,omitempty"`
	KillTicks   []int             `json:"kill_ticks,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
	Victims     []string          `json:"victims,omitempty"`
	Weapon      string            `json:"weapon,omitempty"`
	PlayerSlot  int               `json:"player_slot,omitempty"`
	SteamID     string            `json:"steamid"`
	Demo        string            `json:"demo"`
	SegmentFrom int               `json:"segment_tick_start"`
	SegmentTo   int               `json:"segment_tick_end"`
}

type HighlightResult struct {
	Demo       string      `json:"demo"`
	SteamID    string      `json:"steamid"`
	TickRate   float64     `json:"tick_rate"`
	Highlights []Highlight `json:"highlights"`
}

type ParsedDemo struct {
	Demo     string
	TickRate float64
	Kills    []KillEvent
}

type Player struct {
	SteamID string
	Name    string
	// Team is the side the player was on when the roster was captured
	// (first freezetime): "CT", "T", or "" when unknown.
	Team string
}
