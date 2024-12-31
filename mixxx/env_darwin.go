package mixxx

import (
	"fmt"
	"os"
	"path/filepath"
)

func getDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user homedir: %w", err)
	}
	return filepath.Join(homeDir, "Library", "Containers", "org.mixxx.mixxx", "Data", "Library", "Application Support", "Mixxx", "mixxxdb.sqlite"), nil
}
