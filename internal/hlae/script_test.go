package hlae

import (
	"strings"
	"testing"

	"cs2-demo-highlighter/internal/model"
)

func TestResolveSegmentsMergesOverlap(t *testing.T) {
	builder := NewScriptBuilder()
	builder.StartOffsetTicks = 10
	builder.EndOffsetTicks = 10

	result := model.HighlightResult{
		Highlights: []model.Highlight{
			{SegmentFrom: 100, SegmentTo: 120, Type: model.HighlightWallbang, Round: 1},
			{SegmentFrom: 125, SegmentTo: 140, Type: model.HighlightNoScope, Round: 1},
		},
	}

	segs := builder.resolveSegments(result)
	if len(segs) != 1 {
		t.Fatalf("expected 1 merged segment, got %d", len(segs))
	}
	if segs[0].StartTick != 90 {
		t.Fatalf("unexpected start tick: got %d, want %d", segs[0].StartTick, 90)
	}
	if segs[0].EndTick != 150 {
		t.Fatalf("unexpected end tick: got %d, want %d", segs[0].EndTick, 150)
	}
}

func TestBuildUsesPresetAndPovLock(t *testing.T) {
	builder := NewScriptBuilder()
	builder.FFmpegPreset = "afxFfmpegYuv420p"
	builder.OutputPath = "highlights"
	builder.FrameRate = 120

	result := model.HighlightResult{
		Demo:     "demo.dem",
		SteamID:  "76561197960266727",
		TickRate: 64,
		Highlights: []model.Highlight{
			{
				Type:        model.HighlightWallbang,
				Round:       3,
				PlayerSlot:  7,
				TimeStart:   10,
				TimeEnd:     11,
				SegmentFrom: 1000,
				SegmentTo:   1020,
			},
		},
	}

	script := builder.Build(result)
	if strings.Contains(script, "//") {
		t.Fatalf("script must not contain comments for CS2 console paste")
	}
	if !strings.Contains(script, "mirv_streams settings edit afxDefault settings afxFfmpegYuv420p;") {
		t.Fatalf("missing preset command in script")
	}
	if !strings.Contains(script, "spec_player 7;") {
		t.Fatalf("missing POV lock command")
	}
	if !strings.Contains(script, "demo_resume; wait 120;") {
		t.Fatalf("missing initial resume->wait warmup before first seek")
	}
	if !strings.Contains(script, "demo_pause; demo_gototick 999; spec_player 7; demo_resume;") {
		t.Fatalf("missing initial pause->seek->resume with POV lock")
	}
	if strings.Contains(script, "spec_lock_to_accountid") {
		t.Fatalf("legacy accountid lock must not be present")
	}
	if !strings.Contains(script, "mirv_streams record name highlights_hl_0001_r3_wallbang;") {
		t.Fatalf("expected underscore record name without quoting")
	}
	if !strings.Contains(script, "mirv_deathmsg filter add attackerMatch=!x76561197960266727 block=1 lastRule=1;") {
		t.Fatalf("expected killfeed filter for selected steamid")
	}
	if !strings.Contains(script, "demoui 0;") {
		t.Fatalf("expected demoui command")
	}
	if !strings.Contains(script, "cl_truview_show_status 0;") {
		t.Fatalf("expected TrueView status to be hidden")
	}
	if strings.Contains(script, "startmovie") {
		t.Fatalf("legacy startmovie command must not be present")
	}
}

func TestBuildAddsAutoSkipBetweenSegments(t *testing.T) {
	builder := NewScriptBuilder()

	result := model.HighlightResult{
		Highlights: []model.Highlight{
			{Type: model.HighlightWallbang, Round: 1, PlayerSlot: 4, SegmentFrom: 100, SegmentTo: 120},
			{Type: model.HighlightNoScope, Round: 2, PlayerSlot: 8, SegmentFrom: 300, SegmentTo: 320},
		},
	}

	script := builder.Build(result)
	if !strings.Contains(script, "mirv_cmd addAtTick 121 \"demo_pause; demo_gototick 299; spec_player 8; demo_resume\";") {
		t.Fatalf("expected pause->seek->resume with next segment slot")
	}
	if !strings.Contains(script, "mirv_cmd addAtTick 100 \"spec_player 4; host_framerate 60;") {
		t.Fatalf("expected first segment to use its own slot")
	}
	if !strings.Contains(script, "mirv_cmd addAtTick 322 \"disconnect\";") {
		t.Fatalf("expected disconnect after all regular segments")
	}
}

func TestBuildAddsIntraSegmentJumpForRoundMultikillGap(t *testing.T) {
	builder := NewScriptBuilder()
	builder.StartOffsetTicks = 10
	builder.EndOffsetTicks = 5
	builder.KillGapTicks = 60

	result := model.HighlightResult{
		Highlights: []model.Highlight{
			{
				Type:        model.HighlightMultiKill,
				Round:       10,
				PlayerSlot:  7,
				SegmentFrom: 100,
				SegmentTo:   240,
				KillTicks:   []int{100, 130, 240},
			},
		},
	}

	script := builder.Build(result)
	if !strings.Contains(script, "mirv_cmd addAtTick 136 \"demo_pause; demo_gototick 229; spec_player 7; demo_resume\";") {
		t.Fatalf("expected one intra-segment jump for large kill gap")
	}
	if strings.Count(script, "mirv_streams record start") != 1 {
		t.Fatalf("expected one recording start for the segment")
	}
	if strings.Contains(script, "mirv_cmd addAtTick 106 \"demo_pause; demo_gototick") {
		t.Fatalf("did not expect jump for small kill gap")
	}
}

func TestResolveSegmentsSkipsHeadshotCollectionSummary(t *testing.T) {
	builder := NewScriptBuilder()

	result := model.HighlightResult{
		Highlights: []model.Highlight{
			{Type: model.HighlightHeadshotMix, Round: 1, PlayerSlot: 9, SegmentFrom: 100, SegmentTo: 1000},
			{Type: model.HighlightHeadshot, Round: 1, PlayerSlot: 9, SegmentFrom: 150, SegmentTo: 151},
			{Type: model.HighlightWallbang, Round: 1, PlayerSlot: 9, SegmentFrom: 200, SegmentTo: 210},
		},
	}

	segs := builder.resolveSegments(result)
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment after skipping headshot summary, got %d", len(segs))
	}
	if segs[0].StartTick != 200 || segs[0].EndTick != 210 {
		t.Fatalf("unexpected regular segment range: %d..%d", segs[0].StartTick, segs[0].EndTick)
	}
}

func TestBuildHeadshotMontageSingleOutputFile(t *testing.T) {
	builder := NewScriptBuilder()
	builder.OutputPath = "highlights"
	builder.FrameRate = 120

	result := model.HighlightResult{
		SteamID: "76561197960266727",
		Highlights: []model.Highlight{
			{Type: model.HighlightHeadshot, Round: 1, PlayerSlot: 4, SegmentFrom: 100, SegmentTo: 110},
			{Type: model.HighlightHeadshot, Round: 3, PlayerSlot: 8, SegmentFrom: 300, SegmentTo: 310},
		},
	}

	script := builder.BuildHeadshotMontage(result, "headshot_collection")
	if !strings.Contains(script, "mirv_streams record name highlights_headshot_collection;") {
		t.Fatalf("expected one output file name for headshot montage")
	}
	if strings.Count(script, "mirv_streams record start") != 1 {
		t.Fatalf("expected exactly one record start for montage")
	}
	if strings.Count(script, "mirv_streams record end") < 2 {
		t.Fatalf("expected setup end + final end commands in montage script")
	}
	if !strings.Contains(script, "mirv_cmd addAtTick 111 \"demo_pause; demo_gototick 299; spec_player 8; demo_resume\";") {
		t.Fatalf("expected auto jump between headshot segments")
	}
	if !strings.Contains(script, "mirv_cmd addAtTick 312 \"disconnect\";") {
		t.Fatalf("expected disconnect after headshot montage")
	}
}

func TestHelpersUseStableFallbacks(t *testing.T) {
	builder := NewScriptBuilder()

	if got := headshotMontageNameToken("   "); got != "headshot_collection" {
		t.Fatalf("expected default headshot montage name, got %q", got)
	}
	if got := builder.frameRate(); got != 60 {
		t.Fatalf("expected default framerate 60, got %d", got)
	}
	builder.FrameRate = 120
	if got := builder.frameRate(); got != 120 {
		t.Fatalf("expected custom framerate 120, got %d", got)
	}
	if got := builder.ffmpegPreset(); got != "afxFfmpegYuv420p" {
		t.Fatalf("expected default ffmpeg preset, got %q", got)
	}
	if got := seekTickBefore(0); got != 0 {
		t.Fatalf("expected non-negative seek tick, got %d", got)
	}
	if got := seekTickBefore(8); got != 7 {
		t.Fatalf("expected seek tick 7, got %d", got)
	}
}
