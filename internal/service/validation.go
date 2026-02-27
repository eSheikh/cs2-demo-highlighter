package service

import (
	"errors"
	"fmt"
	"strings"
)

func ValidateSteamID(steamID string) error {
	trimmedSteamID := strings.TrimSpace(steamID)
	if trimmedSteamID == "" {
		return errors.New("steamid is required")
	}

	if len(trimmedSteamID) != 17 {
		return fmt.Errorf("steamid must be a 17-digit steamid64")
	}
	for _, ch := range trimmedSteamID {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("steamid must contain digits only")
		}
	}

	return nil
}
