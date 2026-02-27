package service

import "testing"

func TestValidateSteamID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		steamID string
		wantErr bool
	}{
		{name: "empty", steamID: "", wantErr: true},
		{name: "contains spaces only", steamID: "   ", wantErr: true},
		{name: "short", steamID: "7656119", wantErr: true},
		{name: "contains letters", steamID: "7656119ABCDEFGHIJ", wantErr: true},
		{name: "valid", steamID: "76561197960265728", wantErr: false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateSteamID(tc.steamID)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}
