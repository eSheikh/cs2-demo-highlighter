package demoinfocs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eSheikh/cs2-demo-highlighter/internal/demo"
)

func TestParseRejectsNonDemoExtension(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "match.txt")
	if err := os.WriteFile(path, []byte("payload"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	_, err := NewParser().Parse(context.Background(), path, "76561197960265728")
	if err == nil {
		t.Fatalf("expected error for non-.dem file, got nil")
	}
	if !errors.Is(err, demo.ErrInvalidFileExtension) {
		t.Fatalf("expected %v, got %v", demo.ErrInvalidFileExtension, err)
	}
}

func TestParseHandlesInvalidDemoData(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "broken.dem")
	if err := os.WriteFile(path, []byte("not-a-real-demo"), 0o644); err != nil {
		t.Fatalf("write demo: %v", err)
	}

	_, err := NewParser().Parse(context.Background(), path, "76561197960265728")
	if err == nil {
		t.Fatalf("expected parse error for invalid .dem payload, got nil")
	}
	if !strings.Contains(err.Error(), "demo parsing failed") && !strings.Contains(err.Error(), "truncated or corrupted") {
		t.Fatalf("unexpected parse error: %v", err)
	}
}

func TestParseHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "match.dem")
	if err := os.WriteFile(path, []byte("placeholder-demo"), 0o644); err != nil {
		t.Fatalf("write demo: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewParser().Parse(ctx, path, "76561197960265728")
	if err == nil {
		t.Fatalf("expected context cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
