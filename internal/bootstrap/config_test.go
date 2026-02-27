package bootstrap

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cs2-demo-highlighter/internal/demo"
	"cs2-demo-highlighter/internal/hlae"
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
		"--hlae", "  highlights.cfg  ",
		"--hlae-headshots", "   ",
		"--hlae-headshots-name", "  my_mix  ",
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
	if cfg.HLAE.ScriptPath != "highlights.cfg" {
		t.Fatalf("expected trimmed hlae script path, got %q", cfg.HLAE.ScriptPath)
	}
	if cfg.HLAE.HeadshotMontageScriptPath != "" {
		t.Fatalf("expected empty headshot script path, got %q", cfg.HLAE.HeadshotMontageScriptPath)
	}
	if cfg.HLAE.HeadshotMontageName != "my_mix" {
		t.Fatalf("expected trimmed headshot montage name, got %q", cfg.HLAE.HeadshotMontageName)
	}
	if cfg.HLAE.OutputPath != "clips" {
		t.Fatalf("expected trimmed hlae output path, got %q", cfg.HLAE.OutputPath)
	}
	if cfg.HLAE.FFmpegPreset != "afxFfmpegYuv420p" {
		t.Fatalf("expected trimmed preset, got %q", cfg.HLAE.FFmpegPreset)
	}
}
