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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestArtifactObjectNameUsesOSArchAndVersion(t *testing.T) {
	tests := []struct {
		name    string
		goos    string
		goarch  string
		version string
		want    string
	}{
		{
			name:    "linux",
			goos:    "linux",
			goarch:  "amd64",
			version: "2.4.0",
			want:    "cli/linux/amd64/erda-cli-2.4.0.tar.gz",
		},
		{
			name:    "darwin",
			goos:    "darwin",
			goarch:  "arm64",
			version: "2.4.0-alpha.20260421112000",
			want:    "cli/darwin/arm64/erda-cli-2.4.0-alpha.20260421112000.tar.gz",
		},
		{
			name:    "windows",
			goos:    "windows",
			goarch:  "amd64",
			version: "2.4.0",
			want:    "cli/windows/amd64/erda-cli-2.4.0.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ArtifactObjectName(tt.goos, tt.goarch, tt.version); got != tt.want {
				t.Fatalf("ArtifactObjectName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExecutableFileName(t *testing.T) {
	if got := ExecutableFileName("linux"); got != "erda-cli" {
		t.Fatalf("ExecutableFileName(linux) = %q", got)
	}
	if got := ExecutableFileName("darwin"); got != "erda-cli" {
		t.Fatalf("ExecutableFileName(darwin) = %q", got)
	}
	if got := ExecutableFileName("windows"); got != "erda-cli.exe" {
		t.Fatalf("ExecutableFileName(windows) = %q", got)
	}
}

func TestManifestObjectNames(t *testing.T) {
	if got := VersionManifestObjectName("linux", "amd64", "2.4.0"); got != "cli/linux/amd64/erda-cli-2.4.0.json" {
		t.Fatalf("VersionManifestObjectName() = %q", got)
	}
	if got := ChannelManifestObjectName("darwin", "arm64", ChannelStable); got != "cli/darwin/arm64/stable.json" {
		t.Fatalf("ChannelManifestObjectName() = %q", got)
	}
	if got := ChannelVersionsObjectName("linux", "amd64", ChannelAlpha); got != "cli/linux/amd64/alpha-versions.json" {
		t.Fatalf("ChannelVersionsObjectName() = %q", got)
	}
}

func TestDetectChannelFromVersion(t *testing.T) {
	tests := []struct {
		version string
		want    string
	}{
		{version: "2.4.0", want: ChannelStable},
		{version: "2.4.0-alpha.20260421112000", want: ChannelAlpha},
		{version: "2.4.0-beta.1", want: ChannelBeta},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got, err := DetectChannel(tt.version)
			if err != nil {
				t.Fatalf("DetectChannel() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("DetectChannel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersionComparisonSupportsLegacyMinorOnlyVersion(t *testing.T) {
	newer, err := HasNewerVersion("2.4", "2.4.1")
	if err != nil {
		t.Fatalf("HasNewerVersion() error = %v", err)
	}
	if !newer {
		t.Fatal("expected 2.4.1 to be newer than legacy 2.4")
	}
}

func TestVersionComparisonRejectsDowngradeFromStableToAlpha(t *testing.T) {
	newer, err := HasNewerVersion("2.4.0", "2.4.0-alpha.20260421112000")
	if err != nil {
		t.Fatalf("HasNewerVersion() error = %v", err)
	}
	if newer {
		t.Fatal("stable release should not be considered older than alpha prerelease")
	}
}

func TestUpdateVersionIndexKeepsLatestTenByPublishedAt(t *testing.T) {
	index := &VersionIndex{
		Channel: ChannelAlpha,
		OS:      "linux",
		Arch:    "amd64",
	}
	for i := 0; i < 11; i++ {
		index = UpdateVersionIndex(index, Manifest{
			Version:     fmt.Sprintf("2.4.0-alpha.20260421120%02d", i),
			Channel:     ChannelAlpha,
			OS:          "linux",
			Arch:        "amd64",
			PublishedAt: time.Date(2026, 4, 21, 12, i, 0, 0, time.UTC).Format(time.RFC3339),
		}, 10)
	}

	if len(index.Versions) != 10 {
		t.Fatalf("len(index.Versions) = %d, want 10", len(index.Versions))
	}
	if got := index.Versions[0].PublishedAt; got != time.Date(2026, 4, 21, 12, 10, 0, 0, time.UTC).Format(time.RFC3339) {
		t.Fatalf("latest publishedAt = %q", got)
	}
	if got := index.Versions[len(index.Versions)-1].PublishedAt; got != time.Date(2026, 4, 21, 12, 1, 0, 0, time.UTC).Format(time.RFC3339) {
		t.Fatalf("oldest retained publishedAt = %q", got)
	}
}

func TestUpdateVersionIndexReplacesExistingVersionEntry(t *testing.T) {
	index := &VersionIndex{
		Channel: ChannelStable,
		OS:      "darwin",
		Arch:    "arm64",
		Versions: []Manifest{
			{
				Version:     "2.4.1",
				Channel:     ChannelStable,
				OS:          "darwin",
				Arch:        "arm64",
				PublishedAt: "2026-04-21T10:00:00Z",
				SHA256:      "old",
			},
		},
	}

	index = UpdateVersionIndex(index, Manifest{
		Version:     "2.4.1",
		Channel:     ChannelStable,
		OS:          "darwin",
		Arch:        "arm64",
		PublishedAt: "2026-04-21T11:00:00Z",
		SHA256:      "new",
	}, 10)

	if len(index.Versions) != 1 {
		t.Fatalf("len(index.Versions) = %d, want 1", len(index.Versions))
	}
	if index.Versions[0].SHA256 != "new" {
		t.Fatalf("SHA256 = %q, want new", index.Versions[0].SHA256)
	}
}

func TestFetchVersionIndexReturnsTypedNotFoundErrorOn404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	_, err := FetchVersionIndex(server.URL + "/stable-versions.json")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsVersionIndexNotFound(err) {
		t.Fatalf("expected typed not found error, got %v", err)
	}
}

func TestFetchManifestReturnsTypedNotFoundErrorOn404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	_, err := FetchManifest(server.URL + "/stable.json")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsManifestNotFound(err) {
		t.Fatalf("expected typed not found error, got %v", err)
	}
}

func TestFetchManifestRejectsInvalidURL(t *testing.T) {
	_, err := FetchManifest("file:///tmp/stable.json")
	if err == nil {
		t.Fatal("expected invalid url error")
	}
	if !strings.Contains(err.Error(), "invalid fetch url") {
		t.Fatalf("unexpected error: %v", err)
	}
}
