//go:build mage
// +build mage

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/autonomouskoi/mageutil"
)

func OverlayWebDir() error {
	return mageutil.Mkdir(overlayWebDir)
}

func OverlayWebSrc() {
	mg.SerialDeps(
		OverlayTypeScript,
		OverlayWebSrcCopy,
		OverlayVersion,
		OverlayTSProtos,
	)
}

func OverlayGoProtos() error {
	dest := filepath.Join(overlayDir, "overlay.pb.go")
	src := filepath.Join(overlayDir, "overlay.proto")
	return mageutil.GoProto(dest, src, overlayDir, "module=github.com/autonomouskoi/trackstar/overlay")
}
func OverlayTSProtos() error {
	mg.Deps(OverlayWebDir)
	return mageutil.TSProtosInDir(overlayWebDir, overlayDir)
}

func OverlayVersion() error {
	b, err := json.Marshal(map[string]string{
		"Software": "aktrackstar overlay",
		"Build":    time.Now().Format("20060102-1504"),
	})
	if err != nil {
		return fmt.Errorf("marshalling version: %w", err)
	}
	outPath := filepath.Join(overlayWebDir, "build.json")
	return os.WriteFile(outPath, b, 0644)
}

func OverlayTypeScript() error {
	mg.Deps(OverlayTSProtos)
	return mageutil.BuildTypeScript(overlayDir, overlayDir, overlayWebDir)
}

func OverlayWebSrcCopy() error {
	filenames := []string{"index.html", "ui.html"}
	if err := mageutil.CopyInDir(overlayWebDir, overlayDir, filenames...); err != nil {
		return fmt.Errorf("copying: %w", err)
	}
	return nil
}

func OverlayWebZip() error {
	mg.Deps(OverlayWebSrc)

	zipPath := filepath.Join(overlayDir, "web.zip")
	if err := sh.Rm(zipPath); err != nil {
		return fmt.Errorf("removing %s: %w", zipPath, err)
	}

	return mageutil.ZipDir(overlayWebDir, zipPath)
}
