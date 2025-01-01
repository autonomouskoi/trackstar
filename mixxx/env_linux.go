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
	return filepath.Join(homeDir, ".mixxx", "mixxxdb.sqlite"), nil
}
