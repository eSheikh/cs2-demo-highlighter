package jsonrepo

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cs2-demo-highlighter/internal/model"
)

func TestSaveHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "result.json")
	repo := New(outputPath)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.Save(ctx, model.HighlightResult{Demo: "match.dem"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestSaveTrimsOutputPath(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "result.json")
	repo := New("  " + outputPath + "  ")

	payload := model.HighlightResult{
		Demo:    "match.dem",
		SteamID: "76561197960265728",
	}
	if err := repo.Save(context.Background(), payload); err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var got model.HighlightResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if got.Demo != payload.Demo || got.SteamID != payload.SteamID {
		t.Fatalf("unexpected output payload: %+v", got)
	}
}

func TestSaveSkipsWhitespaceOnlyOutputPath(t *testing.T) {
	t.Parallel()

	repo := New("   ")
	if err := repo.Save(context.Background(), model.HighlightResult{Demo: "match.dem"}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
