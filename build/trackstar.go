//go:build mage
// +build mage

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/mageutil"
)

func TrackstarWebDir() error {
	return mageutil.Mkdir(trackstarWebDir)
}

func TrackstarWebSrc() {
	mg.SerialDeps(
		TrackstarTypeScript,
		TrackstarWebSrcCopy,
		TrackstarVersion,
	)
}

func TrackstarGoProtos() error {
	dest := filepath.Join(trackstarDir, "trackstar.pb.go")
	src := filepath.Join(trackstarDir, "trackstar.proto")
	return mageutil.GoProto(dest, src, trackstarDir, "module=github.com/autonomouskoi/trackstar")
}

func TrackstarTSProtos() error {
	mg.Deps(TrackstarWebDir)
	return mageutil.TSProtosInDir(trackstarWebDir, trackstarDir)
}

func TrackstarVersion() error {
	mg.Deps(TrackstarWebDir)
	b, err := json.Marshal(map[string]string{
		"Software": "aktrackstar",
		"Build":    "v" + akcore.Version,
	})
	if err != nil {
		return fmt.Errorf("marshalling version: %w", err)
	}
	outPath := filepath.Join(trackstarWebDir, "build.json")
	return os.WriteFile(outPath, b, 0644)
}

func TrackstarTypeScript() error {
	mg.Deps(TrackstarWebDir)
	mg.Deps(TrackstarTSProtos)
	return mageutil.BuildTypeScript(trackstarDir, trackstarDir, trackstarWebDir)
}

func TrackstarWebSrcCopy() error {
	mg.Deps(TrackstarWebDir)
	filenames := []string{"index.html"}
	if err := mageutil.CopyInDir(trackstarWebDir, trackstarDir, filenames...); err != nil {
		return fmt.Errorf("copying: %w", err)
	}
	return nil
}

func TrackstarWebZip() error {
	mg.Deps(TrackstarWebSrc)

	zipPath := filepath.Join(trackstarDir, "web.zip")
	if err := sh.Rm(zipPath); err != nil {
		return fmt.Errorf("removing %s: %w", zipPath, err)
	}

	return mageutil.ZipDir(trackstarWebDir, zipPath)
}
