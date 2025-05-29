//go:build mage
// +build mage

package main

import (
	"fmt"
	"path/filepath"

	"github.com/autonomouskoi/mageutil"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func TwitchchatWebDir() error {
	return mageutil.Mkdir(twitchchatWebDir)
}

func TwitchchatWebSrc() {
	mg.SerialDeps(
		TwitchchatTypeScript,
		TwitchchatWebSrcCopy,
	)
}

func TwitchchatGoProtos() error {
	dest := filepath.Join(twitchchatDir, "twitchchat.pb.go")
	src := filepath.Join(twitchchatDir, "twitchchat.proto")
	return mageutil.GoProto(dest, src, twitchchatDir, twitchchatDir, "module=github.com/autonomouskoi/trackstar/twitchchat")
}

func TwitchchatTSProtos() error {
	mg.Deps(TwitchchatWebDir)
	return mageutil.TSProtosInDir(twitchchatWebDir, twitchchatDir, filepath.Join(twitchchatDir, "node_modules"))
}

func TwitchchatTypeScript() error {
	mg.Deps(TwitchchatWebDir)
	mg.Deps(TwitchchatTSProtos)
	return mageutil.BuildTypeScript(twitchchatDir, twitchchatDir, twitchchatWebDir)
}

func TwitchchatWebSrcCopy() error {
	mg.Deps(TwitchchatWebDir)
	filenames := []string{"index.html"}
	if err := mageutil.CopyInDir(twitchchatWebDir, twitchchatDir, filenames...); err != nil {
		return fmt.Errorf("copying: %w", err)
	}
	return nil
}

func TwitchchatWebZip() error {
	mg.Deps(TwitchchatWebSrc)

	zipPath := filepath.Join(twitchchatDir, "web.zip")
	if err := sh.Rm(zipPath); err != nil {
		return fmt.Errorf("removing %s: %w", zipPath, err)
	}

	return mageutil.ZipDir(twitchchatWebDir, zipPath)
}
