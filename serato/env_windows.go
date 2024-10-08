package serato

import (
	"fmt"
	"os"
	"path/filepath"
)

func getSessionsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user home dir: %w", err)
	}
	return filepath.Join(home, "Music", "_Serato_", "History", "Sessions"), nil
}
