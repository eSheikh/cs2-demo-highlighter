package demo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathRequired         = errors.New("demo path is required")
	ErrInvalidFileExtension = errors.New("demo file must have .dem extension")
	ErrNotRegularFile       = errors.New("demo path must point to a regular file")
	ErrEmptyFile            = errors.New("demo file is empty")
)

func ValidatePath(path string) error {
	trimmedPath := strings.TrimSpace(path)
	switch {
	case trimmedPath == "":
		return ErrPathRequired
	case !strings.EqualFold(filepath.Ext(trimmedPath), ".dem"):
		return fmt.Errorf("%w: %q", ErrInvalidFileExtension, trimmedPath)
	}

	fileInfo, err := os.Stat(trimmedPath)
	if err != nil {
		return fmt.Errorf("demo file check failed: %w", err)
	}

	switch {
	case !fileInfo.Mode().IsRegular():
		return fmt.Errorf("%w: %q", ErrNotRegularFile, trimmedPath)
	case fileInfo.Size() == 0:
		return fmt.Errorf("%w: %q", ErrEmptyFile, trimmedPath)
	}

	return nil
}
