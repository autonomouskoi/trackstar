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
	return filepath.Join(home, "AppData", "Local", "VirtualDJ", "History"), nil
}
