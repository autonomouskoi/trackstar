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

func StagelinqWebDir() error {
	return mageutil.Mkdir(stagelinqWebDir)
}

func StagelinqWebSrc() {
	mg.SerialDeps(
		StagelinqTypeScript,
		StagelinqWebSrcCopy,
		StagelinqVersion,
		StagelinqTSProtos,
	)
}

func StagelinqGoProtos() error {
	dest := filepath.Join(stagelinqDir, "stagelinq.pb.go")
	src := filepath.Join(stagelinqDir, "stagelinq.proto")
	return mageutil.GoProto(dest, src, stagelinqDir, "module=github.com/autonomouskoi/trackstar/stagelinq")
}
func StagelinqTSProtos() error {
	mg.Deps(StagelinqWebDir)
	return mageutil.TSProtosInDir(stagelinqWebDir, stagelinqDir)
}

func StagelinqVersion() error {
	b, err := json.Marshal(map[string]string{
		"Software": "aktrackstar stagelinq",
		"Build":    akcore.Version,
	})
	if err != nil {
		return fmt.Errorf("marshalling version: %w", err)
	}
	outPath := filepath.Join(stagelinqWebDir, "build.json")
	return os.WriteFile(outPath, b, 0644)
}

func StagelinqTypeScript() error {
	mg.Deps(StagelinqTSProtos)
	return mageutil.BuildTypeScript(stagelinqDir, stagelinqDir, stagelinqWebDir)
}

func StagelinqWebSrcCopy() error {
	filenames := []string{"index.html"}
	if err := mageutil.CopyInDir(stagelinqWebDir, stagelinqDir, filenames...); err != nil {
		return fmt.Errorf("copying: %w", err)
	}
	return nil
}

func StagelinqWebZip() error {
	mg.Deps(StagelinqWebSrc)

	zipPath := filepath.Join(stagelinqDir, "web.zip")
	if err := sh.Rm(zipPath); err != nil {
		return fmt.Errorf("removing %s: %w", zipPath, err)
	}

	return mageutil.ZipDir(stagelinqWebDir, zipPath)
}
