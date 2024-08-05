//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	baseDir         string
	overlayDir      string
	overlayWebDir   string
	stagelinqDir    string
	stagelinqWebDir string
	trackstarDir    string
	trackstarWebDir string
)

var Default = All

func init() {
	var err error
	baseDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	baseDir = filepath.Join(baseDir, "..")

	overlayDir = filepath.Join(baseDir, "overlay")
	overlayWebDir = filepath.Join(overlayDir, "web")
	stagelinqDir = filepath.Join(baseDir, "stagelinq")
	stagelinqWebDir = filepath.Join(stagelinqDir, "web")
	trackstarDir = baseDir
	trackstarWebDir = filepath.Join(trackstarDir, "web")
}

func Clean() error {
	for _, dir := range []string{
		overlayWebDir,
		overlayWebDir + ".zip",
		stagelinqWebDir,
		stagelinqWebDir + ".zip",
		trackstarWebDir,
		trackstarWebDir + ".zip",
	} {
		if err := sh.Rm(dir); err != nil {
			return fmt.Errorf("removing %s: %w", dir, err)
		}
	}
	return nil
}

func All() {
	mg.Deps(
		OverlayGoProtos,
		OverlayWebZip,
		StagelinqGoProtos,
		StagelinqWebZip,
		TrackstarGoProtos,
		TrackstarWebZip,
	)
}

func Dev() {
	mg.Deps(
		OverlayGoProtos,
		OverlayWebSrc,
		StagelinqGoProtos,
		StagelinqWebSrc,
		TrackstarGoProtos,
		TrackstarWebSrc,
	)
}
