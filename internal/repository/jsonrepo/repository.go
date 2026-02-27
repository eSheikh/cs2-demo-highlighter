package jsonrepo

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

type Repository struct {
	outputPath string
}

func New(outputPath string) *Repository {
	return &Repository{outputPath: strings.TrimSpace(outputPath)}
}

func (r *Repository) Save(ctx context.Context, payload model.HighlightResult) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.outputPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(r.outputPath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.WriteFile(r.outputPath, data, 0o644)
}
