package hlae

import (
	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

// Options holds the shared rendering settings applied to every render target.
type Options struct {
	FrameRate       int
	OutputPath      string
	FFmpegPreset    string
	PreRollSeconds  int
	PostRollSeconds int
	KillGapSeconds  int
}

// Mode selects how a target packages its highlights.
type Mode int

const (
	// ModeClips records each highlight as its own segment (separate takes).
	ModeClips Mode = iota
	// ModeMontage records the selected highlights into one continuous take with jump cuts.
	ModeMontage
)

// Target is a single .cfg to generate: a render mode over a set of highlight
// types. An empty Types selects all types.
type Target struct {
	Mode  Mode
	Types model.Selection
	Path  string
	Name  string
}

func BuildTarget(result model.HighlightResult, options Options, target Target) string {
	builder := configuredBuilder(result, options)
	if target.Mode == ModeMontage {
		return builder.BuildMontage(result, target.Types, target.Name)
	}
	if options.KillGapSeconds > 0 && result.TickRate > 0 {
		builder.KillGapTicks = int(result.TickRate * float64(options.KillGapSeconds))
	}
	return builder.BuildClips(result, target.Types, target.Name)
}

func configuredBuilder(result model.HighlightResult, options Options) *ScriptBuilder {
	builder := NewScriptBuilder()
	builder.ApplyOffsetsSeconds(result.TickRate, options.PreRollSeconds, options.PostRollSeconds)
	builder.FrameRate = options.FrameRate
	builder.OutputPath = options.OutputPath
	builder.FFmpegPreset = options.FFmpegPreset
	return builder
}
