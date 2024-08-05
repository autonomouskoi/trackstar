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
	if err := mageutil.HasExec("protoc"); err != nil {
		return err
	}
	plugin := filepath.Join(trackstarDir, "node_modules/.bin/protoc-gen-es")
	if err := mageutil.HasFiles(plugin); err != nil {
		return err
	}
	protoDestDir := filepath.Join(trackstarWebDir, "pb")
	if err := mageutil.Mkdir(protoDestDir); err != nil {
		return fmt.Errorf("creating %s: %w", protoDestDir, err)
	}
	for _, srcFile := range []string{
		"trackstar.proto",
	} {
		baseName := strings.TrimSuffix(filepath.Base(srcFile), ".proto")
		destFile := filepath.Join(protoDestDir, baseName+"_pb.js")
		srcFile = filepath.Join(trackstarDir, srcFile)
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
			"-I", trackstarDir,
			"--es_out", protoDestDir,
			srcFile,
		)
		if err != nil {
			return fmt.Errorf("generating proto %s -> %s\n: %w", srcFile, destFile, err)
		}
	}
	return nil
}

func TrackstarVersion() error {
	mg.Deps(TrackstarWebDir)
	b, err := json.Marshal(map[string]string{
		"Software": "aktrackstar",
		"Build":    time.Now().Format("20060102-1504"),
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
