package hlae

import (
	"strings"

	"cs2-demo-highlighter/internal/model"
)

type Options struct {
	ScriptPath                string
	HeadshotMontageScriptPath string
	HeadshotMontageName       string
	FrameRate                 int
	OutputPath                string
	FFmpegPreset              string
	PreRollSeconds            int
	PostRollSeconds           int
	KillGapSeconds            int
}

func (o Options) Enabled() bool {
	return strings.TrimSpace(o.ScriptPath) != ""
}

func (o Options) HeadshotMontageEnabled() bool {
	return strings.TrimSpace(o.HeadshotMontageScriptPath) != ""
}

func BuildScript(result model.HighlightResult, options Options) string {
	builder := configuredBuilder(result, options)
	if options.KillGapSeconds > 0 && result.TickRate > 0 {
		builder.KillGapTicks = int(result.TickRate * float64(options.KillGapSeconds))
	}
	return builder.Build(result)
}

func BuildHeadshotMontageScript(result model.HighlightResult, options Options) string {
	builder := configuredBuilder(result, options)
	return builder.BuildHeadshotMontage(result, options.HeadshotMontageName)
}

func configuredBuilder(result model.HighlightResult, options Options) *ScriptBuilder {
	builder := NewScriptBuilder()
	builder.ApplyOffsetsSeconds(result.TickRate, options.PreRollSeconds, options.PostRollSeconds)
	builder.FrameRate = options.FrameRate
	builder.OutputPath = options.OutputPath
	builder.FFmpegPreset = options.FFmpegPreset
	return builder
}
