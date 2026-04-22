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

package release

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

const (
	ChannelStable  = "stable"
	ChannelBeta    = "beta"
	ChannelAlpha   = "alpha"
	DefaultBaseURL = "https://erda-release.oss-cn-hangzhou.aliyuncs.com"
)

var (
	minorOnlyVersionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)$`)
	legacyPreReleasePattern = regexp.MustCompile(`^v?(\d+)\.(\d+)-([0-9A-Za-z.-]+)$`)
	ErrManifestNotFound     = errors.New("release manifest not found")
	ErrVersionIndexNotFound = errors.New("release version index not found")
)

type Manifest struct {
	Version     string `json:"version"`
	Channel     string `json:"channel"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	URL         string `json:"url"`
	SHA256      string `json:"sha256"`
	PublishedAt string `json:"publishedAt,omitempty"`
	BuildTime   string `json:"buildTime,omitempty"`
	CommitID    string `json:"commitID,omitempty"`
}

type VersionIndex struct {
	Channel  string     `json:"channel"`
	OS       string     `json:"os"`
	Arch     string     `json:"arch"`
	Versions []Manifest `json:"versions"`
}

func ArtifactObjectName(goos, goarch, version string) string {
	return path.Join("cli", goos, goarch, artifactFileName(goos, version))
}

func ExecutableFileName(goos string) string {
	if goos == "windows" {
		return "erda-cli.exe"
	}
	return "erda-cli"
}

func VersionManifestObjectName(goos, goarch, version string) string {
	return path.Join("cli", goos, goarch, fmt.Sprintf("erda-cli-%s.json", version))
}

func ChannelManifestObjectName(goos, goarch, channel string) string {
	return path.Join("cli", goos, goarch, fmt.Sprintf("%s.json", channel))
}

func ChannelVersionsObjectName(goos, goarch, channel string) string {
	return path.Join("cli", goos, goarch, fmt.Sprintf("%s-versions.json", channel))
}

func ValidateChannel(channel string) error {
	switch channel {
	case ChannelStable, ChannelBeta, ChannelAlpha:
		return nil
	default:
		return fmt.Errorf("unsupported channel %q", channel)
	}
}

func DetectChannel(version string) (string, error) {
	v, err := parseVersion(version)
	if err != nil {
		return "", err
	}
	pre := strings.ToLower(v.Prerelease())
	switch {
	case pre == "":
		return ChannelStable, nil
	case strings.HasPrefix(pre, ChannelAlpha):
		return ChannelAlpha, nil
	case strings.HasPrefix(pre, ChannelBeta):
		return ChannelBeta, nil
	default:
		return "", fmt.Errorf("unsupported prerelease channel in version %q", version)
	}
}

func HasNewerVersion(currentVersion, candidateVersion string) (bool, error) {
	current, err := parseVersion(currentVersion)
	if err != nil {
		return false, err
	}
	candidate, err := parseVersion(candidateVersion)
	if err != nil {
		return false, err
	}
	return candidate.GreaterThan(current), nil
}

func ChannelManifestURL(baseURL, goos, goarch, channel string) string {
	return objectURL(baseURL, ChannelManifestObjectName(goos, goarch, channel))
}

func ChannelVersionsURL(baseURL, goos, goarch, channel string) string {
	return objectURL(baseURL, ChannelVersionsObjectName(goos, goarch, channel))
}

func VersionManifestURL(baseURL, goos, goarch, version string) string {
	return objectURL(baseURL, VersionManifestObjectName(goos, goarch, version))
}

func FetchManifest(url string) (*Manifest, error) {
	resp, err := fetchURL(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: %s", ErrManifestNotFound, url)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("fetch manifest %s failed: status %d: %s", url, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decode manifest %s failed: %w", url, err)
	}
	return &manifest, nil
}

func IsManifestNotFound(err error) bool {
	return errors.Is(err, ErrManifestNotFound)
}

func FetchVersionIndex(url string) (*VersionIndex, error) {
	resp, err := fetchURL(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: %s", ErrVersionIndexNotFound, url)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("fetch version index %s failed: status %d: %s", url, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var index VersionIndex
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, fmt.Errorf("decode version index %s failed: %w", url, err)
	}
	return &index, nil
}

func IsVersionIndexNotFound(err error) bool {
	return errors.Is(err, ErrVersionIndexNotFound)
}

func fetchURL(rawURL string) (*http.Response, error) {
	parsedURL, err := validateFetchURL(rawURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func validateFetchURL(rawURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid fetch url %q: %w", rawURL, err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("invalid fetch url %q", rawURL)
	}
	return parsedURL, nil
}

func UpdateVersionIndex(index *VersionIndex, manifest Manifest, maxEntries int) *VersionIndex {
	if index == nil {
		index = &VersionIndex{}
	}
	index.Channel = manifest.Channel
	index.OS = manifest.OS
	index.Arch = manifest.Arch

	filtered := make([]Manifest, 0, len(index.Versions)+1)
	for _, item := range index.Versions {
		if item.Version == manifest.Version {
			continue
		}
		filtered = append(filtered, item)
	}
	filtered = append(filtered, manifest)

	sort.SliceStable(filtered, func(i, j int) bool {
		ti := parsePublishedAt(filtered[i].PublishedAt)
		tj := parsePublishedAt(filtered[j].PublishedAt)
		if !ti.Equal(tj) {
			return ti.After(tj)
		}
		newer, err := HasNewerVersion(filtered[j].Version, filtered[i].Version)
		if err == nil && newer {
			return true
		}
		return filtered[i].Version > filtered[j].Version
	})

	if maxEntries > 0 && len(filtered) > maxEntries {
		filtered = filtered[:maxEntries]
	}
	index.Versions = filtered
	return index
}

func parseVersion(raw string) (*semver.Version, error) {
	normalized := normalizeVersion(raw)
	v, err := semver.NewVersion(normalized)
	if err != nil {
		return nil, fmt.Errorf("invalid version %q: %w", raw, err)
	}
	return v, nil
}

func normalizeVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	if matches := minorOnlyVersionPattern.FindStringSubmatch(raw); len(matches) == 3 {
		return fmt.Sprintf("%s.%s.0", matches[1], matches[2])
	}
	if matches := legacyPreReleasePattern.FindStringSubmatch(raw); len(matches) == 4 {
		return fmt.Sprintf("%s.%s.0-%s", matches[1], matches[2], matches[3])
	}
	return strings.TrimPrefix(raw, "v")
}

func artifactFileName(goos, version string) string {
	name := fmt.Sprintf("erda-cli-%s", version)
	if goos == "windows" {
		return name + ".zip"
	}
	return name + ".tar.gz"
}

func objectURL(baseURL, objectName string) string {
	return strings.TrimRight(baseURL, "/") + "/" + objectName
}

func parsePublishedAt(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return t
}
