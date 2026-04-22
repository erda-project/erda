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

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/erda-project/erda/tools/cli/release"
)

func TestParseReleaseArgsRejectsCredentialArgumentsOnCLI(t *testing.T) {
	_, err := parseReleaseArgs([]string{"id", "secret", "linux", "amd64", "2.4.0", "stable", "/tmp/erda-cli"}, nil)
	if err == nil {
		t.Fatal("expected credential arguments to be rejected")
	}

	for _, want := range []string{"environment variables", "ACCESS_KEY_ID", "ACCESS_KEY_SECRET"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to mention %s, got %q", want, err.Error())
		}
	}
}

func TestParseReleaseArgsRejectsMissingCredentialsInEnvironment(t *testing.T) {
	_, err := parseReleaseArgs([]string{"linux", "amd64", "2.4.0", "stable", "/tmp/erda-cli"}, func(string) string { return "" })
	if err == nil {
		t.Fatal("expected missing credential error")
	}

	for _, want := range []string{"ACCESS_KEY_ID", "ACCESS_KEY_SECRET"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to mention %s, got %q", want, err.Error())
		}
	}
}

func TestParseReleaseArgsAcceptsRequiredArguments(t *testing.T) {
	args, err := parseReleaseArgs([]string{"linux", "amd64", "2.4.0", "stable", "/tmp/erda-cli"}, func(key string) string {
		switch key {
		case "ACCESS_KEY_ID":
			return "id"
		case "ACCESS_KEY_SECRET":
			return "secret"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatalf("expected arguments to parse: %v", err)
	}

	if args.keyID != "id" || args.keySecret != "secret" {
		t.Fatalf("unexpected credentials: %+v", args)
	}
	if args.goos != "linux" || args.goarch != "amd64" || args.version != "2.4.0" || args.channel != release.ChannelStable || args.file != "/tmp/erda-cli" {
		t.Fatalf("unexpected parsed args: %+v", args)
	}
}

func TestParsePruneArgsAcceptsDefaults(t *testing.T) {
	args, err := parsePruneArgs(nil, func(key string) string {
		switch key {
		case "ACCESS_KEY_ID":
			return "id"
		case "ACCESS_KEY_SECRET":
			return "secret"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatalf("expected prune args to parse: %v", err)
	}

	if args.keep != 10 {
		t.Fatalf("keep = %d, want 10", args.keep)
	}
	if args.apply {
		t.Fatal("apply should default to false")
	}
	if strings.Join(args.channels, ",") != "alpha,beta" {
		t.Fatalf("channels = %v, want [alpha beta]", args.channels)
	}
}

func TestParsePruneArgsRejectsMissingCredentials(t *testing.T) {
	_, err := parsePruneArgs(nil, func(string) string { return "" })
	if err == nil {
		t.Fatal("expected missing credential error")
	}
	if !strings.Contains(err.Error(), "ACCESS_KEY_ID") || !strings.Contains(err.Error(), "ACCESS_KEY_SECRET") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParsePruneArgsRejectsInvalidKeep(t *testing.T) {
	_, err := parsePruneArgs([]string{"--keep", "0"}, func(key string) string {
		switch key {
		case "ACCESS_KEY_ID":
			return "id"
		case "ACCESS_KEY_SECRET":
			return "secret"
		default:
			return ""
		}
	})
	if err == nil {
		t.Fatal("expected invalid keep error")
	}
	if !strings.Contains(err.Error(), "--keep") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseReleaseArgsRejectsChannelVersionMismatch(t *testing.T) {
	_, err := parseReleaseArgs([]string{"linux", "amd64", "2.4.0-alpha.20260421112000", "stable", "/tmp/erda-cli"}, func(key string) string {
		switch key {
		case "ACCESS_KEY_ID":
			return "id"
		case "ACCESS_KEY_SECRET":
			return "secret"
		default:
			return ""
		}
	})
	if err == nil {
		t.Fatal("expected channel/version mismatch to fail")
	}
	if !strings.Contains(err.Error(), "channel") {
		t.Fatalf("expected channel mismatch error, got %q", err.Error())
	}
}

func TestBuildManifestUsesArchivedArtifactURL(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "erda-cli")
	if err := os.WriteFile(binPath, []byte("cli-binary"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	args := releaseArgs{
		goos:    "linux",
		goarch:  "amd64",
		version: "2.4.0",
		channel: release.ChannelStable,
		file:    binPath,
	}

	archivePath, cleanup, err := packageArtifact(args)
	if err != nil {
		t.Fatalf("packageArtifact() error = %v", err)
	}
	defer cleanup()

	manifest, err := buildManifest(args, archivePath)
	if err != nil {
		t.Fatalf("buildManifest() error = %v", err)
	}

	wantURL := release.DefaultBaseURL + "/cli/linux/amd64/erda-cli-2.4.0.tar.gz"
	if manifest.URL != wantURL {
		t.Fatalf("manifest.URL = %q, want %q", manifest.URL, wantURL)
	}
	if manifest.SHA256 == "" {
		t.Fatal("manifest.SHA256 should not be empty")
	}
}
