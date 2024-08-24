package virtualdj

import (
	"fmt"
	"os"
	"path/filepath"
)

func getHistoryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user home dir: %w", err)
	}
	return filepath.Join(home, "Library", "Application Support", "VirtualDJ", "History"), nil
}
