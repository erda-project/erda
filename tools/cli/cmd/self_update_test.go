// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	infraVersion "github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/release"
	"github.com/erda-project/erda/tools/cli/utils"
)

// saveUpdateDeps saves the package-level function variables used by the update
// commands and restores them when the test finishes.
func saveUpdateDeps(t *testing.T) {
	t.Helper()
	origFetchLatest := fetchLatestReleaseManifest
	origFetchVersion := fetchVersionReleaseManifest
	origVersion := currentCLIVersion
	origApply := applyReleaseUpdate
	origGetConfig := getUpdateGlobalConfig
	origSetConfig := setUpdateGlobalConfig
	origFetchIndex := fetchReleaseVersionIndex
	origStdout := updateStdout
	t.Cleanup(func() {
		fetchLatestReleaseManifest = origFetchLatest
		fetchVersionReleaseManifest = origFetchVersion
		currentCLIVersion = origVersion
		applyReleaseUpdate = origApply
		getUpdateGlobalConfig = origGetConfig
		setUpdateGlobalConfig = origSetConfig
		fetchReleaseVersionIndex = origFetchIndex
		updateStdout = origStdout
	})
}

func TestUpdateCommandShape(t *testing.T) {
	if UPDATE.Name != "update" {
		t.Fatalf("update command name = %q, want update", UPDATE.Name)
	}
	if !hasCommandFlag(UPDATE.Flags, "channel") {
		t.Fatal("update missing --channel")
	}
	if !hasCommandFlag(UPDATE.Flags, "version") {
		t.Fatal("update missing --version")
	}
	if UPDATELIST.ParentName != "UPDATE" {
		t.Fatalf("update list parent = %q, want UPDATE", UPDATELIST.ParentName)
	}
	if UPDATECHECK.ParentName != "UPDATE" {
		t.Fatalf("update check parent = %q, want UPDATE", UPDATECHECK.ParentName)
	}
	if UPDATESETDEFAULT.ParentName != "UPDATE" {
		t.Fatalf("update set-default parent = %q, want UPDATE", UPDATESETDEFAULT.ParentName)
	}
}

func TestUpdateDefaultsToStableChannel(t *testing.T) {
	saveUpdateDeps(t)

	var gotChannel string
	fetchLatestReleaseManifest = func(channel string) (*release.Manifest, error) {
		gotChannel = channel
		return &release.Manifest{Version: "2.4.1", Channel: release.ChannelStable}, nil
	}
	fetchVersionReleaseManifest = func(string) (*release.Manifest, error) {
		t.Fatal("version manifest lookup should not be used")
		return nil, nil
	}
	currentCLIVersion = func() string { return "2.4.0" }
	applyReleaseUpdate = func(*release.Manifest) error { return nil }
	getUpdateGlobalConfig = func() (string, *command.GlobalConfig, error) {
		return "", &command.GlobalConfig{Version: command.ConfigVersion}, utils.NotExist
	}

	if err := UpdateCLI(&command.Context{}, "", ""); err != nil {
		t.Fatalf("UpdateCLI() error = %v", err)
	}
	if gotChannel != release.ChannelStable {
		t.Fatalf("channel = %q, want %q", gotChannel, release.ChannelStable)
	}
}

func TestUpdateUsesConfiguredDefaultChannel(t *testing.T) {
	saveUpdateDeps(t)

	var gotChannel string
	fetchLatestReleaseManifest = func(channel string) (*release.Manifest, error) {
		gotChannel = channel
		return &release.Manifest{Version: "2.4.1", Channel: channel}, nil
	}
	fetchVersionReleaseManifest = func(string) (*release.Manifest, error) {
		t.Fatal("version manifest lookup should not be used")
		return nil, nil
	}
	currentCLIVersion = func() string { return "2.4.0" }
	applyReleaseUpdate = func(*release.Manifest) error { return nil }
	getUpdateGlobalConfig = func() (string, *command.GlobalConfig, error) {
		return "/tmp/config", &command.GlobalConfig{Version: command.ConfigVersion, UpdateChannel: release.ChannelAlpha}, nil
	}

	if err := UpdateCLI(&command.Context{}, "", ""); err != nil {
		t.Fatalf("UpdateCLI() error = %v", err)
	}
	if gotChannel != release.ChannelAlpha {
		t.Fatalf("channel = %q, want %q", gotChannel, release.ChannelAlpha)
	}
}

func TestUpdateSupportsExplicitVersion(t *testing.T) {
	saveUpdateDeps(t)

	var gotVersion string
	fetchLatestReleaseManifest = func(string) (*release.Manifest, error) {
		t.Fatal("latest manifest lookup should not be used when --version is set")
		return nil, nil
	}
	fetchVersionReleaseManifest = func(version string) (*release.Manifest, error) {
		gotVersion = version
		return &release.Manifest{Version: version, Channel: release.ChannelStable}, nil
	}
	currentCLIVersion = func() string { return "2.4.0" }
	applyReleaseUpdate = func(*release.Manifest) error { return nil }

	if err := UpdateCLI(&command.Context{}, "", "2.4.1"); err != nil {
		t.Fatalf("UpdateCLI() error = %v", err)
	}
	if gotVersion != "2.4.1" {
		t.Fatalf("version = %q, want 2.4.1", gotVersion)
	}
}

func TestUpdateSupportsExplicitOlderVersion(t *testing.T) {
	saveUpdateDeps(t)

	fetchLatestReleaseManifest = func(string) (*release.Manifest, error) {
		t.Fatal("latest manifest lookup should not be used when --version is set")
		return nil, nil
	}
	fetchVersionReleaseManifest = func(version string) (*release.Manifest, error) {
		return &release.Manifest{Version: version, Channel: release.ChannelStable}, nil
	}
	currentCLIVersion = func() string { return "2.4.2" }

	called := false
	applyReleaseUpdate = func(*release.Manifest) error {
		called = true
		return nil
	}

	if err := UpdateCLI(&command.Context{}, "", "2.4.1"); err != nil {
		t.Fatalf("UpdateCLI() error = %v", err)
	}
	if !called {
		t.Fatal("expected explicit older version to be applied")
	}
}

func TestUpdateSkipsWhenAlreadyOnLatestVersion(t *testing.T) {
	saveUpdateDeps(t)

	fetchLatestReleaseManifest = func(channel string) (*release.Manifest, error) {
		return &release.Manifest{Version: "2.4.0", Channel: channel}, nil
	}
	currentCLIVersion = func() string { return "2.4.0" }
	applyReleaseUpdate = func(*release.Manifest) error {
		t.Fatal("applyReleaseUpdate should not be called when already on latest version")
		return nil
	}

	if err := UpdateCLI(&command.Context{}, release.ChannelStable, ""); err != nil {
		t.Fatalf("UpdateCLI() error = %v", err)
	}
}

func TestUpdateAllowsLegacyEmbeddedVersionComparison(t *testing.T) {
	saveUpdateDeps(t)

	fetchLatestReleaseManifest = func(channel string) (*release.Manifest, error) {
		return &release.Manifest{Version: "2.4.1", Channel: channel}, nil
	}
	currentCLIVersion = func() string { return "2.4" }

	called := false
	applyReleaseUpdate = func(*release.Manifest) error {
		called = true
		return nil
	}

	if err := UpdateCLI(&command.Context{}, release.ChannelStable, ""); err != nil {
		t.Fatalf("UpdateCLI() error = %v", err)
	}
	if !called {
		t.Fatal("expected update to be applied for legacy 2.4 version")
	}
}

func TestUpdateCurrentVersionDefaultsToEmbeddedVersion(t *testing.T) {
	orig := infraVersion.Version
	t.Cleanup(func() { infraVersion.Version = orig })
	infraVersion.Version = "2.4.0"

	if got := defaultCurrentCLIVersion(); got != "2.4.0" {
		t.Fatalf("defaultCurrentCLIVersion() = %q, want 2.4.0", got)
	}
}

func TestUpdateListDefaultsToStableChannel(t *testing.T) {
	saveUpdateDeps(t)

	var gotChannel string
	fetchReleaseVersionIndex = func(channel string) (*release.VersionIndex, error) {
		gotChannel = channel
		return &release.VersionIndex{
			Channel: channel,
			OS:      "darwin",
			Arch:    "arm64",
			Versions: []release.Manifest{
				{Version: "2.4.1", Channel: channel, PublishedAt: "2026-04-21T10:00:00Z"},
			},
		}, nil
	}

	var out bytes.Buffer
	updateStdout = &out
	getUpdateGlobalConfig = func() (string, *command.GlobalConfig, error) {
		return "", &command.GlobalConfig{Version: command.ConfigVersion}, utils.NotExist
	}

	if err := UpdateList(&command.Context{}, ""); err != nil {
		t.Fatalf("UpdateList() error = %v", err)
	}
	if gotChannel != release.ChannelStable {
		t.Fatalf("channel = %q, want %q", gotChannel, release.ChannelStable)
	}
	if !bytes.Contains(out.Bytes(), []byte("2.4.1")) {
		t.Fatalf("output = %q, want version row", out.String())
	}
}

func TestUpdateSetDefaultWritesGlobalConfig(t *testing.T) {
	saveUpdateDeps(t)

	getUpdateGlobalConfig = func() (string, *command.GlobalConfig, error) {
		return "/tmp/config", &command.GlobalConfig{Version: command.ConfigVersion, Host: "https://erda.cloud"}, nil
	}

	var gotFile string
	var gotConfig *command.GlobalConfig
	setUpdateGlobalConfig = func(file string, conf *command.GlobalConfig) error {
		gotFile = file
		gotConfig = conf
		return nil
	}

	if err := UpdateSetDefault(&command.Context{}, release.ChannelAlpha); err != nil {
		t.Fatalf("UpdateSetDefault() error = %v", err)
	}
	if gotFile != "/tmp/config" {
		t.Fatalf("config file = %q, want /tmp/config", gotFile)
	}
	if gotConfig.UpdateChannel != release.ChannelAlpha {
		t.Fatalf("update channel = %q, want %q", gotConfig.UpdateChannel, release.ChannelAlpha)
	}
	if gotConfig.Host != "https://erda.cloud" {
		t.Fatalf("host should be preserved, got %q", gotConfig.Host)
	}
}

func TestUpdateSetDefaultRequiresChannel(t *testing.T) {
	err := UpdateSetDefault(&command.Context{}, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "please specify default update channel") {
		t.Fatalf("unexpected error = %q", err.Error())
	}
}

func TestUpdateListReturnsFriendlyMessageWhenVersionIndexMissing(t *testing.T) {
	saveUpdateDeps(t)

	fetchReleaseVersionIndex = func(channel string) (*release.VersionIndex, error) {
		return nil, fmt.Errorf("%w: %s", release.ErrVersionIndexNotFound, "https://example.com/stable-versions.json")
	}

	err := UpdateList(&command.Context{}, release.ChannelStable)
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, release.ErrVersionIndexNotFound) {
		t.Fatalf("expected friendly user-facing error, got sentinel error: %v", err)
	}
	if !strings.Contains(err.Error(), "no remote versions found") {
		t.Fatalf("unexpected error = %q", err.Error())
	}
}

func TestUpdateCheckDoesNotApplyUpdate(t *testing.T) {
	saveUpdateDeps(t)

	fetchLatestReleaseManifest = func(channel string) (*release.Manifest, error) {
		return &release.Manifest{Version: "2.4.1", Channel: channel}, nil
	}
	currentCLIVersion = func() string { return "2.4.0" }
	applyReleaseUpdate = func(*release.Manifest) error {
		t.Fatal("applyReleaseUpdate should not be called by update check")
		return nil
	}

	if err := CheckUpdate(&command.Context{}, release.ChannelStable, ""); err != nil {
		t.Fatalf("CheckUpdate() error = %v", err)
	}
}

func TestUpdateReturnsFriendlyMessageWhenChannelManifestMissing(t *testing.T) {
	saveUpdateDeps(t)

	fetchLatestReleaseManifest = func(channel string) (*release.Manifest, error) {
		return nil, fmt.Errorf("%w: %s", release.ErrManifestNotFound, "https://example.com/stable.json")
	}
	fetchVersionReleaseManifest = func(string) (*release.Manifest, error) {
		t.Fatal("version manifest lookup should not be used")
		return nil, nil
	}

	err := UpdateCLI(&command.Context{}, release.ChannelStable, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, release.ErrManifestNotFound) {
		t.Fatalf("expected friendly user-facing error, got sentinel error: %v", err)
	}
	if !strings.Contains(err.Error(), `no remote release found for channel "stable"`) {
		t.Fatalf("unexpected error = %q", err.Error())
	}
}

func TestUpdateReturnsFriendlyMessageWhenVersionManifestMissing(t *testing.T) {
	saveUpdateDeps(t)

	fetchLatestReleaseManifest = func(string) (*release.Manifest, error) {
		t.Fatal("latest manifest lookup should not be used")
		return nil, nil
	}
	fetchVersionReleaseManifest = func(version string) (*release.Manifest, error) {
		return nil, fmt.Errorf("%w: %s", release.ErrManifestNotFound, "https://example.com/erda-cli-"+version+".json")
	}

	err := UpdateCLI(&command.Context{}, "", "2.4.1")
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, release.ErrManifestNotFound) {
		t.Fatalf("expected friendly user-facing error, got sentinel error: %v", err)
	}
	if !strings.Contains(err.Error(), `no remote release found for version "2.4.1"`) {
		t.Fatalf("unexpected error = %q", err.Error())
	}
}

func TestExtractReleaseExecutableFromTarGz(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "erda-cli.tar.gz")
	wantContent := []byte("linux-cli")

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	header := &tar.Header{
		Name: release.ExecutableFileName("linux"),
		Mode: 0o755,
		Size: int64(len(wantContent)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("WriteHeader() error = %v", err)
	}
	if _, err := tw.Write(wantContent); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar.Close() error = %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip.Close() error = %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("file.Close() error = %v", err)
	}

	dstPath := filepath.Join(tmpDir, "erda-cli")
	manifest := &release.Manifest{OS: "linux", URL: "https://example.com/erda-cli-2.4.0.tar.gz"}
	if err := extractReleaseExecutable(manifest, archivePath, dstPath); err != nil {
		t.Fatalf("extractReleaseExecutable() error = %v", err)
	}
	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(got, wantContent) {
		t.Fatalf("extracted content = %q, want %q", string(got), string(wantContent))
	}
}

func TestExtractReleaseExecutableFromZip(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "erda-cli.zip")
	wantContent := []byte("windows-cli")

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create(release.ExecutableFileName("windows"))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := w.Write(wantContent); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip.Close() error = %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("file.Close() error = %v", err)
	}

	dstPath := filepath.Join(tmpDir, "erda-cli.exe")
	manifest := &release.Manifest{OS: "windows", URL: "https://example.com/erda-cli-2.4.0.zip"}
	if err := extractReleaseExecutable(manifest, archivePath, dstPath); err != nil {
		t.Fatalf("extractReleaseExecutable() error = %v", err)
	}
	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(got, wantContent) {
		t.Fatalf("extracted content = %q, want %q", string(got), string(wantContent))
	}
}

func TestDefaultApplyReleaseUpdateRejectsInvalidReleaseURL(t *testing.T) {
	err := defaultApplyReleaseUpdate(&release.Manifest{
		URL: "file:///tmp/erda-cli.tar.gz",
	})
	if err == nil {
		t.Fatal("expected invalid release url error")
	}
	if !strings.Contains(err.Error(), "invalid release url") {
		t.Fatalf("unexpected error = %q", err.Error())
	}
}

func TestStageReplacementExecutableCopiesBinaryIntoTargetDir(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src-erda-cli")
	targetPath := filepath.Join(tmpDir, "bin", "erda-cli")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	want := []byte("updated-cli")
	if err := os.WriteFile(srcPath, want, 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stagedPath, err := stageReplacementExecutable(srcPath, targetPath, 0o755)
	if err != nil {
		t.Fatalf("stageReplacementExecutable() error = %v", err)
	}
	defer os.Remove(stagedPath)

	if filepath.Dir(stagedPath) != filepath.Dir(targetPath) {
		t.Fatalf("staged path dir = %q, want %q", filepath.Dir(stagedPath), filepath.Dir(targetPath))
	}
	got, err := os.ReadFile(stagedPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("staged content = %q, want %q", string(got), string(want))
	}
}
