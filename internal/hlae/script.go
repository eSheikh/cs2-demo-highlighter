package hlae

import (
	"cmp"
	"fmt"
	"sort"
	"strings"

	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

const (
	defaultFrameRate = 60
	defaultPreset    = "afxFfmpegYuv420p"
)

// ScriptBuilder generates HLAE commands for CS2 highlight recording.
//
// Output is intentionally command-only (no comment lines), because CS2/HLAE
// console paste can ignore or break on comment-heavy blocks.
type ScriptBuilder struct {
	StartOffsetTicks int
	EndOffsetTicks   int
	KillGapTicks     int
	FrameRate        int
	OutputPath       string
	FFmpegPreset     string
}

func NewScriptBuilder() *ScriptBuilder {
	return &ScriptBuilder{
		StartOffsetTicks: 0,
		EndOffsetTicks:   0,
		KillGapTicks:     0,
		FrameRate:        defaultFrameRate,
		OutputPath:       "",
		FFmpegPreset:     defaultPreset,
	}
}

func (b *ScriptBuilder) ApplyOffsetsSeconds(tickRate float64, preRollSeconds int, postRollSeconds int) {
	if tickRate <= 0 {
		return
	}
	if preRollSeconds > 0 {
		b.StartOffsetTicks = int(tickRate * float64(preRollSeconds))
	}
	if postRollSeconds > 0 {
		b.EndOffsetTicks = int(tickRate * float64(postRollSeconds))
	}
}

type recordingSegment struct {
	Index      int
	Name       string
	PlayerSlot int
	StartTick  int
	EndTick    int
	Highlights []model.Highlight
}

type segmentRange struct {
	Highlight model.Highlight
	StartTick int
	EndTick   int
}

type recordingJump struct {
	AtTick     int
	SeekTick   int
	PlayerSlot int
}

func (b *ScriptBuilder) Build(result model.HighlightResult) string {
	var w strings.Builder
	segs := b.resolveSegments(result)

	b.writeSetup(&w, result.SteamID)
	b.writeTickCommands(&w, segs)
	b.writeFooter(&w, segs)

	return w.String()
}

func (b *ScriptBuilder) BuildHeadshotMontage(result model.HighlightResult, montageName string) string {
	var w strings.Builder
	segs := b.resolveHeadshotSegments(result)

	b.writeSetup(&w, result.SteamID)
	b.writeHeadshotMontageCommands(&w, segs, montageName)
	b.writeHeadshotMontageFooter(&w, segs, montageName)

	return w.String()
}

func (b *ScriptBuilder) resolveSegments(result model.HighlightResult) []recordingSegment {
	return b.resolveSegmentsFromHighlights(result.Highlights, shouldSkipDefaultRecording)
}

func (b *ScriptBuilder) resolveHeadshotSegments(result model.HighlightResult) []recordingSegment {
	return b.resolveSegmentsFromHighlights(result.Highlights, shouldSkipHeadshotMontage)
}

func (b *ScriptBuilder) resolveSegmentsFromHighlights(highlights []model.Highlight, shouldSkip func(model.HighlightType) bool) []recordingSegment {
	ranges := make([]segmentRange, 0, len(highlights))
	for _, h := range highlights {
		if shouldSkip != nil && shouldSkip(h.Type) {
			continue
		}
		start := max(h.SegmentFrom-b.StartOffsetTicks, 0)
		end := max(h.SegmentTo+b.EndOffsetTicks, start)
		ranges = append(ranges, segmentRange{
			Highlight: h,
			StartTick: start,
			EndTick:   end,
		})
	}

	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].StartTick == ranges[j].StartTick {
			return ranges[i].EndTick < ranges[j].EndTick
		}
		return ranges[i].StartTick < ranges[j].StartTick
	})

	segments := make([]recordingSegment, 0, len(ranges))
	for _, item := range ranges {
		if len(segments) == 0 {
			segments = append(segments, recordingSegment{
				PlayerSlot: item.Highlight.PlayerSlot,
				StartTick:  item.StartTick,
				EndTick:    item.EndTick,
				Highlights: []model.Highlight{item.Highlight},
			})
			continue
		}

		last := &segments[len(segments)-1]
		if item.StartTick <= last.EndTick {
			last.EndTick = max(last.EndTick, item.EndTick)
			if last.PlayerSlot <= 0 && item.Highlight.PlayerSlot > 0 {
				last.PlayerSlot = item.Highlight.PlayerSlot
			}
			last.Highlights = append(last.Highlights, item.Highlight)
			continue
		}

		segments = append(segments, recordingSegment{
			PlayerSlot: item.Highlight.PlayerSlot,
			StartTick:  item.StartTick,
			EndTick:    item.EndTick,
			Highlights: []model.Highlight{item.Highlight},
		})
	}

	for i := range segments {
		segments[i].Index = i + 1
		segments[i].Name = b.buildName(segments[i])
	}

	return segments
}

func shouldSkipDefaultRecording(highlightType model.HighlightType) bool {
	return highlightType == model.HighlightHeadshotMix || highlightType == model.HighlightHeadshot
}

func shouldSkipHeadshotMontage(highlightType model.HighlightType) bool {
	return highlightType != model.HighlightHeadshot
}

func (b *ScriptBuilder) writeSetup(w *strings.Builder, steamID string) {
	writeCommandLine(w, "mirv_cvar_unhide_all")
	writeCommandLine(w, "mirv_cmd clear")
	writeCommandLine(w, "mirv_streams record end")
	writeCommandLine(w, fmt.Sprintf("mirv_streams settings edit afxDefault settings %s", b.ffmpegPreset()))
	writeCommandLine(w, "mirv_streams record screen enabled 1")
	writeCommandLine(w, fmt.Sprintf("mirv_streams record fps %d", b.frameRate()))
	writeCommandLine(w, "spec_show_xray 0")
	writeCommandLine(w, "demoui 0")
	writeCommandLine(w, "cl_truview_show_status 0")
	writeCommandLine(w, "cl_drawhud 0")
	writeCommandLine(w, "cl_drawhud_force_radar -1")
	writeCommandLine(w, "cl_drawhud_force_deathnotices 1")
	writeCommandLine(w, "mirv_deathmsg filter clear")
	if steamID != "" {
		writeCommandLine(w, fmt.Sprintf("mirv_deathmsg localPlayer x%s", steamID))
		writeCommandLine(w, fmt.Sprintf("mirv_deathmsg filter add attackerMatch=!x%s block=1 lastRule=1", steamID))
	}
	writeCommandLine(w, "toggleconsole")
	w.WriteString("\n")
}

func (b *ScriptBuilder) writeTickCommands(w *strings.Builder, segs []recordingSegment) {
	if len(segs) == 0 {
		writeCommandLine(w, "echo \"No highlights found.\"")
		return
	}

	for i, seg := range segs {
		recordPath := b.resolveRecordPath(seg.Name)
		startParts := append(povCommandsBySlot(seg.PlayerSlot),
			fmt.Sprintf("host_framerate %d", b.frameRate()),
			fmt.Sprintf("mirv_streams record name %s", recordPath),
			"mirv_streams record start",
		)
		startCmd := joinCommands(startParts...)
		endCmd := joinCommands("mirv_streams record end", "host_framerate 0")

		writeCommandLine(w, fmt.Sprintf("mirv_cmd addAtTick %d \"%s\"", seg.StartTick, escapeForAddAtTick(startCmd)))
		writeCommandLine(w, fmt.Sprintf("mirv_cmd addAtTick %d \"%s\"", seg.EndTick, escapeForAddAtTick(endCmd)))
		for _, jump := range b.resolveIntraSegmentJumps(seg) {
			jumpParts := append([]string{"demo_pause", fmt.Sprintf("demo_gototick %d", jump.SeekTick)}, povCommandsBySlot(jump.PlayerSlot)...)
			jumpParts = append(jumpParts, "demo_resume")
			writeCommandLine(w, fmt.Sprintf("mirv_cmd addAtTick %d \"%s\"", jump.AtTick, escapeForAddAtTick(joinCommands(jumpParts...))))
		}

		if i+1 < len(segs) {
			next := segs[i+1]
			nextSeek := seekTickBefore(next.StartTick)
			writeCommandLine(w, fmt.Sprintf("mirv_cmd addAtTick %d \"%s\"", seg.EndTick+1, escapeForAddAtTick(buildSeekJumpCommand(nextSeek, next.PlayerSlot))))
		}
	}

	lastTick := segs[len(segs)-1].EndTick + 1
	writeCommandLine(w, fmt.Sprintf("mirv_cmd addAtTick %d \"echo === All %d segments recorded ===\"", lastTick, len(segs)))
	writeCommandLine(w, fmt.Sprintf("mirv_cmd addAtTick %d \"disconnect\"", lastTick+1))
	w.WriteString("\n")

	b.writeInitialSeek(w, segs[0], "Auto-seek to first segment", true)
}

func (b *ScriptBuilder) writeHeadshotMontageCommands(w *strings.Builder, segs []recordingSegment, montageName string) {
	if len(segs) == 0 {
		writeCommandLine(w, "echo \"No headshot highlights found.\"")
		return
	}

	first := segs[0]
	recordPath := b.resolveRecordPath(headshotMontageNameToken(montageName))

	startParts := append(povCommandsBySlot(first.PlayerSlot),
		fmt.Sprintf("host_framerate %d", b.frameRate()),
		fmt.Sprintf("mirv_streams record name %s", recordPath),
		"mirv_streams record start",
	)
	writeCommandLine(w, fmt.Sprintf("mirv_cmd addAtTick %d \"%s\"", first.StartTick, escapeForAddAtTick(joinCommands(startParts...))))

	for i, seg := range segs {
		if i+1 >= len(segs) {
			continue
		}

		next := segs[i+1]
		nextSeek := seekTickBefore(next.StartTick)
		writeCommandLine(w, fmt.Sprintf("mirv_cmd addAtTick %d \"%s\"", seg.EndTick+1, escapeForAddAtTick(buildSeekJumpCommand(nextSeek, next.PlayerSlot))))
	}

	last := segs[len(segs)-1]
	stopCmd := joinCommands("mirv_streams record end", "host_framerate 0")
	writeCommandLine(w, fmt.Sprintf("mirv_cmd addAtTick %d \"%s\"", last.EndTick, escapeForAddAtTick(stopCmd)))
	writeCommandLine(w, fmt.Sprintf("mirv_cmd addAtTick %d \"echo === Headshot montage recorded ===\"", last.EndTick+1))
	writeCommandLine(w, fmt.Sprintf("mirv_cmd addAtTick %d \"disconnect\"", last.EndTick+2))
	w.WriteString("\n")

	b.writeInitialSeek(w, first, "Auto-seek to first headshot segment", true)
}

func (b *ScriptBuilder) writeHeadshotMontageFooter(w *strings.Builder, segs []recordingSegment, montageName string) {
	if len(segs) == 0 {
		writeCommandLine(w, "echo \"Loaded 0 headshot montage segments.\"")
		return
	}
	nameToken := headshotMontageNameToken(montageName)
	writeCommandLine(w, fmt.Sprintf("echo \"Loaded %d headshot montage segments into one recording.\"",
		len(segs)))
	writeCommandLine(w, fmt.Sprintf("echo \"Output name: %s\"", b.resolveRecordPath(nameToken)))
}

func headshotMontageNameToken(rawName string) string {
	return cmp.Or(sanitizeNameToken(rawName), "headshot_collection")
}

func (b *ScriptBuilder) writeFooter(w *strings.Builder, segs []recordingSegment) {
	if len(segs) == 0 {
		writeCommandLine(w, "echo \"Loaded 0 segments.\"")
		return
	}
	writeCommandLine(w, fmt.Sprintf("echo \"Loaded %d recording segments via mirv_cmd.\"", len(segs)))
	writeCommandLine(w, "echo \"POV lock mode: spec_player by saved slot per highlight.\"")
	writeCommandLine(w, fmt.Sprintf("echo \"Auto-skip enabled. First start tick: %d.\"", segs[0].StartTick))
}

func (b *ScriptBuilder) resolveIntraSegmentJumps(seg recordingSegment) []recordingJump {
	if b.KillGapTicks <= 0 {
		return nil
	}

	jumpByTick := make(map[int]recordingJump)
	for _, h := range seg.Highlights {
		if h.Type != model.HighlightMultiKill || len(h.KillTicks) < 2 {
			continue
		}
		for i := 0; i+1 < len(h.KillTicks); i++ {
			prevTick := h.KillTicks[i]
			nextTick := h.KillTicks[i+1]
			if nextTick-prevTick < b.KillGapTicks {
				continue
			}

			jumpAt := max(prevTick+b.EndOffsetTicks+1, seg.StartTick+1)
			if jumpAt >= seg.EndTick {
				continue
			}

			nextWindowStart := max(nextTick-b.StartOffsetTicks, 0)
			seekTick := seekTickBefore(nextWindowStart)
			if seekTick <= jumpAt {
				continue
			}

			slot := seg.PlayerSlot
			if h.PlayerSlot > 0 {
				slot = h.PlayerSlot
			}

			existing, exists := jumpByTick[jumpAt]
			if !exists || seekTick > existing.SeekTick {
				jumpByTick[jumpAt] = recordingJump{
					AtTick:     jumpAt,
					SeekTick:   seekTick,
					PlayerSlot: slot,
				}
			}
		}
	}

	jumps := make([]recordingJump, 0, len(jumpByTick))
	for _, jump := range jumpByTick {
		jumps = append(jumps, jump)
	}
	sort.Slice(jumps, func(i, j int) bool {
		if jumps[i].AtTick == jumps[j].AtTick {
			return jumps[i].SeekTick < jumps[j].SeekTick
		}
		return jumps[i].AtTick < jumps[j].AtTick
	})

	filtered := make([]recordingJump, 0, len(jumps))
	lastSeekTick := -1
	for _, jump := range jumps {
		if jump.AtTick <= lastSeekTick || jump.SeekTick <= lastSeekTick {
			continue
		}
		filtered = append(filtered, jump)
		lastSeekTick = jump.SeekTick
	}

	return filtered
}

func (b *ScriptBuilder) writeInitialSeek(w *strings.Builder, seg recordingSegment, label string, withWarmup bool) {
	firstSeek := seekTickBefore(seg.StartTick)
	writeCommandLine(w, fmt.Sprintf("echo \"%s: tick %d\"", label, firstSeek))
	if withWarmup {
		writeCommandLine(w, joinCommands("demo_resume", fmt.Sprintf("wait %d", b.frameRate())))
	}
	writeCommandLine(w, buildSeekJumpCommand(firstSeek, seg.PlayerSlot))
	w.WriteString("\n")
}

func (b *ScriptBuilder) frameRate() int {
	return max(b.FrameRate, defaultFrameRate)
}

func (b *ScriptBuilder) ffmpegPreset() string {
	return cmp.Or(sanitizePresetToken(b.FFmpegPreset), defaultPreset)
}

func (b *ScriptBuilder) resolveRecordPath(segmentName string) string {
	base := sanitizeNameToken(strings.ReplaceAll(strings.TrimSpace(b.OutputPath), "/", "_"))
	if base == "" {
		return segmentName
	}
	return base + "_" + segmentName
}

func (b *ScriptBuilder) buildName(seg recordingSegment) string {
	if len(seg.Highlights) == 0 {
		return fmt.Sprintf("hl_%04d", seg.Index)
	}
	first := seg.Highlights[0]
	typeToken := sanitizeNameToken(string(first.Type))
	if typeToken == "" {
		typeToken = "highlight"
	}
	if len(seg.Highlights) > 1 {
		typeToken = "cluster_" + typeToken
	}
	return fmt.Sprintf("hl_%04d_r%d_%s", seg.Index, first.Round, typeToken)
}

func sanitizeNameToken(value string) string {
	if value == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		".", "_",
		"-", "_",
	)
	clean := replacer.Replace(strings.ToLower(strings.TrimSpace(value)))
	buf := make([]rune, 0, len(clean))
	for _, r := range clean {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			buf = append(buf, r)
		}
	}
	return string(buf)
}

func sanitizePresetToken(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	buf := make([]rune, 0, len(trimmed))
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			buf = append(buf, r)
		}
	}
	return string(buf)
}

func escapeForAddAtTick(command string) string {
	command = strings.ReplaceAll(command, `\`, `\\`)
	command = strings.ReplaceAll(command, `"`, `\"`)
	return command
}

func seekTickBefore(tick int) int {
	return max(tick-1, 0)
}

func povCommandsBySlot(playerSlot int) []string {
	if playerSlot <= 0 {
		return nil
	}
	return []string{
		fmt.Sprintf("spec_player %d", playerSlot),
	}
}

func buildSeekJumpCommand(seekTick int, playerSlot int) string {
	parts := append([]string{"demo_pause", fmt.Sprintf("demo_gototick %d", seekTick)}, povCommandsBySlot(playerSlot)...)
	parts = append(parts, "demo_resume")
	return joinCommands(parts...)
}

func joinCommands(commands ...string) string {
	parts := make([]string, 0, len(commands))
	for _, cmd := range commands {
		trimmed := strings.TrimSpace(cmd)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, "; ")
}

func writeCommandLine(w *strings.Builder, command string) {
	trimmed := strings.TrimSpace(command)
	if strings.HasSuffix(trimmed, ";") {
		w.WriteString(trimmed)
		w.WriteString("\n")
		return
	}
	w.WriteString(trimmed)
	w.WriteString(";\n")
}
