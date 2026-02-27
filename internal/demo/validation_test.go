package demo

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	validDemo := filepath.Join(tempDir, "match.dem")
	if err := os.WriteFile(validDemo, []byte("demo-content"), 0o644); err != nil {
		t.Fatalf("write valid demo: %v", err)
	}

	emptyDemo := filepath.Join(tempDir, "empty.dem")
	if err := os.WriteFile(emptyDemo, []byte{}, 0o644); err != nil {
		t.Fatalf("write empty demo: %v", err)
	}

	demoDir := filepath.Join(tempDir, "folder.dem")
	if err := os.MkdirAll(demoDir, 0o755); err != nil {
		t.Fatalf("create demo directory: %v", err)
	}

	testCases := []struct {
		name      string
		path      string
		expectErr error
		wantErr   bool
	}{
		{name: "empty path", path: "", expectErr: ErrPathRequired},
		{name: "wrong extension", path: filepath.Join(tempDir, "match.txt"), expectErr: ErrInvalidFileExtension},
		{name: "not existing file", path: filepath.Join(tempDir, "missing.dem"), wantErr: true},
		{name: "directory path", path: demoDir, expectErr: ErrNotRegularFile},
		{name: "empty file", path: emptyDemo, expectErr: ErrEmptyFile},
		{name: "valid demo", path: validDemo},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePath(tc.path)
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
