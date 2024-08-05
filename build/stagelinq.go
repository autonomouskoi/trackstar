//go:build mage
// +build mage

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"

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
	if err := mageutil.HasExec("protoc"); err != nil {
		return err
	}
	plugin := filepath.Join(stagelinqDir, "node_modules/.bin/protoc-gen-es")
	if err := mageutil.HasFiles(plugin); err != nil {
		return err
	}
	protoDestDir := filepath.Join(stagelinqWebDir, "pb")
	if err := mageutil.Mkdir(protoDestDir); err != nil {
		return fmt.Errorf("creating %s: %w", protoDestDir, err)
	}
	for _, srcFile := range []string{
		"stagelinq.proto",
	} {
		baseName := strings.TrimSuffix(filepath.Base(srcFile), ".proto")
		destFile := filepath.Join(protoDestDir, baseName+"_pb.js")
		srcFile = filepath.Join(stagelinqDir, srcFile)
		newer, err := target.Path(destFile, srcFile)
		if err != nil {
			return fmt.Errorf("testing %s vs %s: %w", srcFile, destFile, err)
		}
		if !newer {
			continue
		}
		mageutil.VerboseF("generating proto %s -> %s\n", srcFile, destFile)
		err = sh.Run("protoc",
			"--plugin", "protoc-gen-es="+plugin+".cmd",
			"-I", stagelinqDir,
			"--es_out", protoDestDir,
			srcFile,
		)
		if err != nil {
			return fmt.Errorf("generating proto %s -> %s\n: %w", srcFile, destFile, err)
		}
	}
	return nil
}

func StagelinqVersion() error {
	b, err := json.Marshal(map[string]string{
		"Software": "aktrackstar stagelinq",
		"Build":    time.Now().Format("20060102-1504"),
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
