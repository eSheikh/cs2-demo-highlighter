package bootstrap

import (
	"flag"
	"fmt"
	"strings"

	"github.com/eSheikh/cs2-demo-highlighter/internal/demo"
	"github.com/eSheikh/cs2-demo-highlighter/internal/hlae"
	"github.com/eSheikh/cs2-demo-highlighter/internal/service"
)

type Config struct {
	DemoPath   string
	SteamID    string
	OutputPath string
	HLAE       hlae.Options
}

func ParseConfig(args []string) (Config, error) {
	cfg := Config{
		OutputPath: "highlights.json",
		HLAE: hlae.Options{
			ScriptPath:                "highlights.cfg",
			HeadshotMontageScriptPath: "headshots.cfg",
			HeadshotMontageName:       "headshot_collection",
			FrameRate:                 60,
			OutputPath:                "highlights",
			FFmpegPreset:              "afxFfmpegYuv420p",
			PreRollSeconds:            3,
			PostRollSeconds:           2,
			KillGapSeconds:            10,
		},
	}

	flags := flag.NewFlagSet("highlighter", flag.ContinueOnError)
	flags.StringVar(&cfg.DemoPath, "demo", "", "path to .dem file")
	flags.StringVar(&cfg.SteamID, "steamid", "", "steamid64 to filter kills")
	flags.StringVar(&cfg.OutputPath, "out", cfg.OutputPath, "output json path")
	flags.StringVar(&cfg.HLAE.ScriptPath, "hlae", cfg.HLAE.ScriptPath, "output HLAE automation script path")
	flags.StringVar(&cfg.HLAE.HeadshotMontageScriptPath, "hlae-headshots", cfg.HLAE.HeadshotMontageScriptPath, "output HLAE script path for one-file headshot montage")
	flags.StringVar(&cfg.HLAE.HeadshotMontageName, "hlae-headshots-name", cfg.HLAE.HeadshotMontageName, "recording output name for headshot montage")
	flags.IntVar(&cfg.HLAE.FrameRate, "hlae-fps", cfg.HLAE.FrameRate, "recording framerate")
	flags.StringVar(&cfg.HLAE.OutputPath, "hlae-path", cfg.HLAE.OutputPath, "recording output folder prefix used by mirv_streams record name")
	flags.StringVar(&cfg.HLAE.FFmpegPreset, "hlae-preset", cfg.HLAE.FFmpegPreset, "HLAE ffmpeg preset for mirv_streams")
	flags.IntVar(&cfg.HLAE.PreRollSeconds, "hlae-preroll", cfg.HLAE.PreRollSeconds, "seconds added before each highlight")
	flags.IntVar(&cfg.HLAE.PostRollSeconds, "hlae-postroll", cfg.HLAE.PostRollSeconds, "seconds added after each highlight")
	flags.IntVar(&cfg.HLAE.KillGapSeconds, "hlae-kill-gap", cfg.HLAE.KillGapSeconds, "seconds between kills in round_multikill to trigger in-recording gototick jump (0 disables)")

	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) normalize() {
	c.DemoPath = strings.TrimSpace(c.DemoPath)
	c.SteamID = strings.TrimSpace(c.SteamID)
	c.OutputPath = strings.TrimSpace(c.OutputPath)

	c.HLAE.ScriptPath = strings.TrimSpace(c.HLAE.ScriptPath)
	c.HLAE.HeadshotMontageScriptPath = strings.TrimSpace(c.HLAE.HeadshotMontageScriptPath)
	c.HLAE.HeadshotMontageName = strings.TrimSpace(c.HLAE.HeadshotMontageName)
	c.HLAE.OutputPath = strings.TrimSpace(c.HLAE.OutputPath)
	c.HLAE.FFmpegPreset = strings.TrimSpace(c.HLAE.FFmpegPreset)
}

func (c Config) Validate() error {
	if err := service.ValidateSteamID(c.SteamID); err != nil {
		return err
	}
	if err := demo.ValidatePath(c.DemoPath); err != nil {
		return err
	}

	for _, check := range []struct {
		flag  string
		value int
	}{
		{flag: "hlae-preroll", value: c.HLAE.PreRollSeconds},
		{flag: "hlae-postroll", value: c.HLAE.PostRollSeconds},
		{flag: "hlae-kill-gap", value: c.HLAE.KillGapSeconds},
	} {
		if check.value < 0 {
			return fmt.Errorf("%s must be >= 0", check.flag)
		}
	}

	return nil
}
