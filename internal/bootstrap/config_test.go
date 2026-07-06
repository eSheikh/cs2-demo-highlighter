package bootstrap

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/eSheikh/cs2-demo-highlighter/internal/demo"
	"github.com/eSheikh/cs2-demo-highlighter/internal/hlae"
	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

func TestConfigValidateDemoPathAndSteamID(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	validDemo := filepath.Join(tempDir, "valid.dem")
	if err := os.WriteFile(validDemo, []byte("demo-content"), 0o644); err != nil {
		t.Fatalf("write valid demo: %v", err)
	}

	testCases := []struct {
		name      string
		config    Config
		expectErr error
		wantErr   bool
	}{
		{
			name: "missing demo path",
			config: Config{
				SteamID: "76561197960265728",
				HLAE:    hlae.Options{},
			},
			expectErr: demo.ErrPathRequired,
		},
		{
			name: "invalid extension",
			config: Config{
				DemoPath: filepath.Join(tempDir, "match.txt"),
				SteamID:  "76561197960265728",
				HLAE:     hlae.Options{},
			},
			expectErr: demo.ErrInvalidFileExtension,
		},
		{
			name: "invalid steamid",
			config: Config{
				DemoPath: validDemo,
				SteamID:  "7656119",
				HLAE:     hlae.Options{},
			},
			wantErr: true,
		},
		{
			name: "valid config",
			config: Config{
				DemoPath: validDemo,
				SteamID:  "76561197960265728",
				HLAE:     hlae.Options{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()
			if tc.expectErr == nil && !tc.wantErr {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			if tc.expectErr == nil && tc.wantErr {
				if err == nil {
					t.Fatalf("expected validation error, got nil")
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error %v, got nil", tc.expectErr)
			}
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected error %v, got %v", tc.expectErr, err)
			}
		})
	}
}

func TestParseConfigTrimsCLIInputs(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	validDemo := filepath.Join(tempDir, "valid.dem")
	if err := os.WriteFile(validDemo, []byte("demo-content"), 0o644); err != nil {
		t.Fatalf("write valid demo: %v", err)
	}

	cfg, err := ParseConfig([]string{
		"--demo", "  " + validDemo + "  ",
		"--steamid", " 76561197960265728 ",
		"--out", "  output.json  ",
		"--clips", "clutch_win,wallbang=clips.cfg",
		"--montage", "headshot_kill=hs.cfg",
		"--hlae-path", "  clips  ",
		"--hlae-preset", "  afxFfmpegYuv420p  ",
	})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if cfg.DemoPath != validDemo {
		t.Fatalf("expected trimmed demo path %q, got %q", validDemo, cfg.DemoPath)
	}
	if cfg.SteamID != "76561197960265728" {
		t.Fatalf("expected trimmed steamid, got %q", cfg.SteamID)
	}
	if cfg.OutputPath != "output.json" {
		t.Fatalf("expected trimmed output path, got %q", cfg.OutputPath)
	}
	if len(cfg.Renders) != 2 {
		t.Fatalf("expected 2 render targets, got %d", len(cfg.Renders))
	}
	clips := cfg.Renders[0]
	if clips.Mode != hlae.ModeClips || clips.Path != "clips.cfg" || clips.Name != "clips" {
		t.Fatalf("unexpected clips target: %+v", clips)
	}
	if !clips.Types[model.HighlightClutchWin] || !clips.Types[model.HighlightWallbang] {
		t.Fatalf("unexpected clips types: %+v", clips.Types)
	}
	montage := cfg.Renders[1]
	if montage.Mode != hlae.ModeMontage || montage.Path != "hs.cfg" || montage.Name != "hs" {
		t.Fatalf("unexpected montage target: %+v", montage)
	}
	if !montage.Types[model.HighlightHeadshot] {
		t.Fatalf("unexpected montage types: %+v", montage.Types)
	}
	if cfg.HLAE.OutputPath != "clips" {
		t.Fatalf("expected trimmed hlae output path, got %q", cfg.HLAE.OutputPath)
	}
	if cfg.HLAE.FFmpegPreset != "afxFfmpegYuv420p" {
		t.Fatalf("expected trimmed preset, got %q", cfg.HLAE.FFmpegPreset)
	}
}

func TestParseConfigDefaultsToClipsRender(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	validDemo := filepath.Join(tempDir, "valid.dem")
	if err := os.WriteFile(validDemo, []byte("demo-content"), 0o644); err != nil {
		t.Fatalf("write valid demo: %v", err)
	}

	cfg, err := ParseConfig([]string{"--demo", validDemo, "--steamid", "76561197960265728"})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if len(cfg.Renders) != 1 {
		t.Fatalf("expected 1 default render, got %d", len(cfg.Renders))
	}
	target := cfg.Renders[0]
	if target.Mode != hlae.ModeClips || target.Path != "highlights.cfg" || len(target.Types) != 0 {
		t.Fatalf("unexpected default render: %+v", target)
	}
}

func TestParseConfigCustomOutputPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	validDemo := filepath.Join(tempDir, "valid.dem")
	if err := os.WriteFile(validDemo, []byte("demo-content"), 0o644); err != nil {
		t.Fatalf("write valid demo: %v", err)
	}

	cfg, err := ParseConfig([]string{
		"--demo", validDemo,
		"--steamid", "76561197960265728",
		"--hlae-path", "C:\\recordings",
	})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if cfg.HLAE.OutputPath != "C:\\recordings" {
		t.Fatalf("expected output path %q, got %q", "C:\\recordings", cfg.HLAE.OutputPath)
	}
}

func TestParseConfigDefaultOutputPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	validDemo := filepath.Join(tempDir, "valid.dem")
	if err := os.WriteFile(validDemo, []byte("demo-content"), 0o644); err != nil {
		t.Fatalf("write valid demo: %v", err)
	}

	cfg, err := ParseConfig([]string{
		"--demo", validDemo,
		"--steamid", "76561197960265728",
	})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if cfg.HLAE.OutputPath == "" {
		t.Fatalf("expected non-empty default output path")
	}
	cwd, _ := os.Getwd()
	expected := filepath.Clean(cwd)
	if cfg.HLAE.OutputPath != expected {
		t.Fatalf("expected default output path %q, got %q", expected, cfg.HLAE.OutputPath)
	}
}
