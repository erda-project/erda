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
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"

	infraVersion "github.com/erda-project/erda-infra/base/version"

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/release"
	"github.com/erda-project/erda/tools/cli/utils"
)

var UPDATE = command.Command{
	Name:      "update",
	ShortHelp: "update erda-cli from official OSS releases",
	Example: ` $ erda-cli update
  $ erda-cli update --channel alpha
  $ erda-cli update --version 2.4.1`,
	Flags: []command.Flag{
		command.StringFlag{Name: "channel", Doc: "release channel: stable, beta, alpha; uses configured default when omitted"},
		command.StringFlag{Name: "version", Doc: "specific version to install"},
	},
	Run: UpdateCLI,
}

var (
	fetchLatestReleaseManifest            = defaultFetchLatestReleaseManifest
	fetchVersionReleaseManifest           = defaultFetchVersionReleaseManifest
	fetchReleaseVersionIndex              = defaultFetchReleaseVersionIndex
	currentCLIVersion                     = defaultCurrentCLIVersion
	applyReleaseUpdate                    = defaultApplyReleaseUpdate
	getUpdateGlobalConfig                 = command.GetGlobalConfig
	setUpdateGlobalConfig                 = command.SetGlobalConfig
	findUpdateGlobalConfig                = utils.FindGlobalConfig
	updateHTTPClient                      = &http.Client{Timeout: 60 * time.Second}
	updateStdout                io.Writer = os.Stdout
)

func UpdateCLI(ctx *command.Context, channel, version string) error {
	manifest, currentVersion, newer, err := resolveUpdateTarget(channel, version)
	if err != nil {
		return err
	}

	if version == "" {
		if !newer {
			ctx.Succ("erda-cli is already up to date (%s)", currentVersion)
			return nil
		}
	} else {
		if manifest.Version == currentVersion {
			ctx.Succ("erda-cli is already on version %s", currentVersion)
			return nil
		}
	}

	if err := applyReleaseUpdate(manifest); err != nil {
		return err
	}

	ctx.Succ("updated erda-cli from %s to %s", currentVersion, manifest.Version)
	return nil
}

func CheckUpdate(ctx *command.Context, channel, version string) error {
	manifest, currentVersion, newer, err := resolveUpdateTarget(channel, version)
	if err != nil {
		return err
	}

	if version == "" {
		if !newer {
			ctx.Succ("erda-cli is already up to date (%s)", currentVersion)
			return nil
		}
		ctx.Info("update available: %s -> %s (%s)", currentVersion, manifest.Version, manifest.Channel)
		return nil
	}

	if manifest.Version == currentVersion {
		ctx.Succ("erda-cli is already on version %s", currentVersion)
		return nil
	}
	if newer {
		ctx.Info("requested update available: %s -> %s", currentVersion, manifest.Version)
	} else {
		ctx.Info("requested version is available: %s -> %s", currentVersion, manifest.Version)
	}
	return nil
}

func defaultFetchLatestReleaseManifest(channel string) (*release.Manifest, error) {
	goos, goarch := currentPlatform()
	return release.FetchManifest(release.ChannelManifestURL(release.DefaultBaseURL, goos, goarch, channel))
}

func defaultFetchVersionReleaseManifest(version string) (*release.Manifest, error) {
	goos, goarch := currentPlatform()
	return release.FetchManifest(release.VersionManifestURL(release.DefaultBaseURL, goos, goarch, version))
}

func defaultFetchReleaseVersionIndex(channel string) (*release.VersionIndex, error) {
	goos, goarch := currentPlatform()
	return release.FetchVersionIndex(release.ChannelVersionsURL(release.DefaultBaseURL, goos, goarch, channel))
}

func defaultCurrentCLIVersion() string {
	return infraVersion.Version
}

func defaultApplyReleaseUpdate(manifest *release.Manifest) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("update is not supported on windows yet")
	}

	targetPath, err := os.Executable()
	if err != nil {
		return err
	}

	parsedURL, err := url.Parse(manifest.URL)
	if err != nil {
		return fmt.Errorf("invalid release url %q: %w", manifest.URL, err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid release url %q", manifest.URL)
	}

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := updateHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s failed: status %d", manifest.URL, resp.StatusCode)
	}

	tmpDir, err := os.MkdirTemp("", "erda-cli-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	archiveFile, err := os.CreateTemp(tmpDir, "archive-*")
	if err != nil {
		return err
	}
	archivePath := archiveFile.Name()
	defer os.Remove(archivePath)

	sum := sha256.New()
	if _, err := io.Copy(io.MultiWriter(archiveFile, sum), resp.Body); err != nil {
		archiveFile.Close()
		return err
	}
	if err := archiveFile.Close(); err != nil {
		return err
	}

	if manifest.SHA256 != "" {
		got := fmt.Sprintf("%x", sum.Sum(nil))
		if !strings.EqualFold(got, manifest.SHA256) {
			return fmt.Errorf("downloaded file checksum mismatch: got %s want %s", got, manifest.SHA256)
		}
	}

	tmpFile, err := os.CreateTemp(tmpDir, "bin-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	if err := extractReleaseExecutable(manifest, archivePath, tmpPath); err != nil {
		return err
	}

	mode := os.FileMode(0o755)
	if info, err := os.Stat(targetPath); err == nil {
		mode = info.Mode()
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		return err
	}

	stagedPath, err := stageReplacementExecutable(tmpPath, targetPath, mode)
	if err != nil {
		return err
	}
	defer os.Remove(stagedPath)

	if err := os.Rename(stagedPath, targetPath); err != nil {
		return updateReplaceError(targetPath, err)
	}
	return nil
}

func stageReplacementExecutable(srcPath, targetPath string, mode os.FileMode) (string, error) {
	stageDir := filepath.Dir(targetPath)
	stagedFile, err := os.CreateTemp(stageDir, ".erda-cli-update-*")
	if err != nil {
		return "", updateReplaceError(targetPath, err)
	}
	stagedPath := stagedFile.Name()
	if err := stagedFile.Close(); err != nil {
		os.Remove(stagedPath)
		return "", err
	}

	keep := false
	defer func() {
		if !keep {
			os.Remove(stagedPath)
		}
	}()

	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.OpenFile(stagedPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return "", updateReplaceError(targetPath, err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return "", err
	}
	if err := dst.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(stagedPath, mode); err != nil {
		return "", err
	}
	keep = true
	return stagedPath, nil
}

func updateReplaceError(targetPath string, err error) error {
	if os.IsPermission(err) {
		return fmt.Errorf("failed to replace %s: %w; reinstall erda-cli into a user-writable directory or rerun with sufficient permissions", targetPath, err)
	}
	return err
}

func extractReleaseExecutable(manifest *release.Manifest, archivePath, dstPath string) error {
	switch {
	case strings.HasSuffix(manifest.URL, ".zip"):
		return extractZipExecutable(manifest.OS, archivePath, dstPath)
	case strings.HasSuffix(manifest.URL, ".tar.gz"):
		return extractTarGzExecutable(manifest.OS, archivePath, dstPath)
	default:
		return fmt.Errorf("unsupported release archive format: %s", manifest.URL)
	}
}

func extractTarGzExecutable(goos, archivePath, dstPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	wantName := release.ExecutableFileName(goos)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if filepath.Base(header.Name) != wantName {
			continue
		}
		return writeExecutableFromReader(tr, dstPath)
	}
	return fmt.Errorf("executable %s not found in archive", wantName)
}

func extractZipExecutable(goos, archivePath, dstPath string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zr.Close()

	wantName := release.ExecutableFileName(goos)
	for _, f := range zr.File {
		if filepath.Base(f.Name) != wantName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		return writeExecutableFromReader(rc, dstPath)
	}
	return fmt.Errorf("executable %s not found in archive", wantName)
}

func writeExecutableFromReader(r io.Reader, dstPath string) error {
	out, err := os.OpenFile(dstPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	closed := false
	defer func() {
		if !closed {
			out.Close()
			os.Remove(dstPath)
		}
	}()
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	closed = true
	return out.Close()
}

func currentPlatform() (string, string) {
	return runtime.GOOS, runtime.GOARCH
}

func UpdateList(ctx *command.Context, channel string) error {
	channel, err := resolveUpdateChannel(channel)
	if err != nil {
		return err
	}

	index, err := fetchReleaseVersionIndex(channel)
	if err != nil {
		if release.IsVersionIndexNotFound(err) {
			goos, goarch := currentPlatform()
			return fmt.Errorf("no remote versions found for channel %q on %s/%s yet", channel, goos, goarch)
		}
		return err
	}

	return writeUpdateIndex(updateStdout, index)
}

func resolveUpdateTarget(channel, version string) (*release.Manifest, string, bool, error) {
	if strings.TrimSpace(version) == "" {
		resolvedChannel, err := resolveUpdateChannel(channel)
		if err != nil {
			return nil, "", false, err
		}
		channel = resolvedChannel
	} else {
		version = strings.TrimSpace(version)
	}

	var (
		manifest *release.Manifest
		err      error
	)
	if version != "" {
		manifest, err = fetchVersionReleaseManifest(version)
	} else {
		manifest, err = fetchLatestReleaseManifest(channel)
	}
	if err != nil {
		if release.IsManifestNotFound(err) {
			goos, goarch := currentPlatform()
			if version != "" {
				return nil, "", false, fmt.Errorf("no remote release found for version %q on %s/%s", version, goos, goarch)
			}
			return nil, "", false, fmt.Errorf("no remote release found for channel %q on %s/%s yet", channel, goos, goarch)
		}
		return nil, "", false, err
	}

	currentVersion := currentCLIVersion()
	newer, err := release.HasNewerVersion(currentVersion, manifest.Version)
	if err != nil {
		return nil, "", false, err
	}
	return manifest, currentVersion, newer, nil
}

func UpdateSetDefault(ctx *command.Context, channel string) error {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return fmt.Errorf("please specify default update channel: stable, beta, alpha")
	}
	if err := release.ValidateChannel(channel); err != nil {
		return err
	}

	configFile, globalConfig, err := getUpdateGlobalConfig()
	if err != nil && err != utils.NotExist {
		return err
	}
	if err == utils.NotExist {
		configFile, err = findUpdateGlobalConfig()
		if err != nil && err != utils.NotExist {
			return err
		}
		globalConfig = &command.GlobalConfig{Version: command.ConfigVersion}
	}
	if globalConfig.Version == "" {
		globalConfig.Version = command.ConfigVersion
	}
	globalConfig.UpdateChannel = channel

	if err := setUpdateGlobalConfig(configFile, globalConfig); err != nil {
		return err
	}
	ctx.Succ("default update channel set to %s", channel)
	return nil
}

func resolveUpdateChannel(channel string) (string, error) {
	channel = strings.TrimSpace(channel)
	if channel != "" {
		if err := release.ValidateChannel(channel); err != nil {
			return "", err
		}
		return channel, nil
	}

	_, globalConfig, err := getUpdateGlobalConfig()
	if err != nil && err != utils.NotExist {
		return "", err
	}
	if err == utils.NotExist || globalConfig == nil || strings.TrimSpace(globalConfig.UpdateChannel) == "" {
		return release.ChannelStable, nil
	}

	configuredChannel := strings.TrimSpace(globalConfig.UpdateChannel)
	if err := release.ValidateChannel(configuredChannel); err != nil {
		return "", fmt.Errorf("invalid default update channel %q in global config: %w", configuredChannel, err)
	}
	return configuredChannel, nil
}

func writeUpdateIndex(w io.Writer, index *release.VersionIndex) error {
	if index == nil {
		return fmt.Errorf("release index is empty")
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintf(tw, "CHANNEL\tOS\tARCH\tVERSION\tPUBLISHEDAT\n"); err != nil {
		return err
	}
	for _, item := range index.Versions {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", index.Channel, index.OS, index.Arch, item.Version, item.PublishedAt); err != nil {
			return err
		}
	}
	return tw.Flush()
}
