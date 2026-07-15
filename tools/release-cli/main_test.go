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
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/erda-project/erda/tools/cli/release"
)

func TestCompletePublishArgsRejectsMissingCredentials(t *testing.T) {
	err := completePublishOptions(&publishOptions{
		version: "2.4.0",
		dir:     "/tmp/bin",
	})
	if err == nil {
		t.Fatal("expected missing credential error")
	}
	for _, want := range []string{"ACCESS_KEY_ID", "ACCESS_KEY_SECRET"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to mention %s, got %q", want, err.Error())
		}
	}
}

func TestCompletePublishArgsDerivesChannelFromVersion(t *testing.T) {
	args := publishOptions{
		keyID:      "id",
		keySecret:  "secret",
		version:    "2.4.0-alpha.20260421112000",
		dir:        "/tmp/bin",
		endpoint:   defaultOSSEndpoint,
		bucketName: defaultOSSBucketName,
	}

	if err := completePublishOptions(&args); err != nil {
		t.Fatalf("expected publish args to complete: %v", err)
	}
	if args.channel != release.ChannelAlpha {
		t.Fatalf("channel = %q, want %q", args.channel, release.ChannelAlpha)
	}
}

func TestResolvePublishReleasesUsesEmbeddedArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	version := "2.4.0"
	for _, artifact := range releaseTargets {
		path := filepath.Join(tmpDir, artifact.fileName)
		if err := os.WriteFile(path, []byte("cli-binary-"+version), 0o755); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", path, err)
		}
	}

	releases, err := resolvePublishTargets(publishOptions{
		keyID:      "id",
		keySecret:  "secret",
		version:    version,
		channel:    release.ChannelStable,
		dir:        tmpDir,
		endpoint:   defaultOSSEndpoint,
		bucketName: defaultOSSBucketName,
	})
	if err != nil {
		t.Fatalf("expected embedded artifact resolution to succeed: %v", err)
	}
	if len(releases) != len(releaseTargets) {
		t.Fatalf("resolved %d releases, want %d", len(releases), len(releaseTargets))
	}
	if releases[0].goos != "darwin" || releases[0].goarch != "arm64" {
		t.Fatalf("unexpected first release target: %+v", releases[0])
	}
	if releases[1].goos != "darwin" || releases[1].goarch != "amd64" {
		t.Fatalf("unexpected second release target: %+v", releases[1])
	}
	if releases[2].goos != "linux" || releases[2].goarch != "amd64" {
		t.Fatalf("unexpected third release target: %+v", releases[2])
	}
}

func TestResolvePublishReleasesRejectsVersionMismatchedArtifact(t *testing.T) {
	tmpDir := t.TempDir()
	for _, artifact := range releaseTargets {
		path := filepath.Join(tmpDir, artifact.fileName)
		if err := os.WriteFile(path, []byte("cli-binary-2.4.0-alpha.20260428153026"), 0o755); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", path, err)
		}
	}

	_, err := resolvePublishTargets(publishOptions{
		keyID:      "id",
		keySecret:  "secret",
		version:    "2.4.0-alpha.20260428153040",
		channel:    release.ChannelAlpha,
		dir:        tmpDir,
		endpoint:   defaultOSSEndpoint,
		bucketName: defaultOSSBucketName,
	})
	if err == nil {
		t.Fatal("expected version mismatch artifact error")
	}
	if !strings.Contains(err.Error(), "does not contain expected version") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolvePublishReleasesRejectsMissingArtifactFile(t *testing.T) {
	_, err := resolvePublishTargets(publishOptions{
		keyID:      "id",
		keySecret:  "secret",
		version:    "2.4.0",
		channel:    release.ChannelStable,
		dir:        t.TempDir(),
		endpoint:   defaultOSSEndpoint,
		bucketName: defaultOSSBucketName,
	})
	if err == nil {
		t.Fatal("expected missing artifact file error")
	}
	if !strings.Contains(err.Error(), "erda-cli") {
		t.Fatalf("expected missing artifact file in error, got %q", err.Error())
	}
}

func TestParsePruneArgsAcceptsDefaults(t *testing.T) {
	args := pruneOptions{keep: 10}
	fillPruneCredentials(&args, func(key string) string {
		if key == "ACCESS_KEY_ID" {
			return "id"
		}
		if key == "ACCESS_KEY_SECRET" {
			return "secret"
		}
		return ""
	})
	err := completePruneOptions(&args, "alpha,beta")
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
	args := pruneOptions{keep: 10}
	fillPruneCredentials(&args, func(string) string { return "" })
	err := completePruneOptions(&args, "alpha,beta")
	if err == nil {
		t.Fatal("expected missing credential error")
	}
	if !strings.Contains(err.Error(), "ACCESS_KEY_ID") || !strings.Contains(err.Error(), "ACCESS_KEY_SECRET") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParsePruneArgsRejectsInvalidKeep(t *testing.T) {
	args := pruneOptions{keep: 0}
	fillPruneCredentials(&args, func(key string) string {
		if key == "ACCESS_KEY_ID" {
			return "id"
		}
		if key == "ACCESS_KEY_SECRET" {
			return "secret"
		}
		return ""
	})
	err := completePruneOptions(&args, "alpha,beta")
	if err == nil {
		t.Fatal("expected invalid keep error")
	}
	if !strings.Contains(err.Error(), "--keep") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompletePublishArgsRejectsChannelVersionMismatch(t *testing.T) {
	err := completePublishOptions(&publishOptions{
		keyID:      "id",
		keySecret:  "secret",
		version:    "2.4.0-alpha.20260421112000",
		channel:    release.ChannelStable,
		dir:        "/tmp/bin",
		endpoint:   defaultOSSEndpoint,
		bucketName: defaultOSSBucketName,
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

	args := publishTarget{
		baseURL: "https://oss.example.com/custom-bucket",
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

	wantURL := "https://oss.example.com/custom-bucket/cli/linux/amd64/erda-cli-2.4.0.tar.gz"
	if manifest.URL != wantURL {
		t.Fatalf("manifest.URL = %q, want %q", manifest.URL, wantURL)
	}
	if manifest.SHA256 == "" {
		t.Fatal("manifest.SHA256 should not be empty")
	}
}

func TestFillPublishCredentialsBaseURLPriority(t *testing.T) {
	t.Run("flag overrides env", func(t *testing.T) {
		opts := publishOptions{baseURL: "https://flag.example.com/release"}
		fillPublishCredentials(&opts, func(key string) string {
			if key == "OSS_BASE_URL" {
				return "https://env.example.com/release"
			}
			return ""
		})
		if opts.baseURL != "https://flag.example.com/release" {
			t.Fatalf("baseURL = %q, want flag value", opts.baseURL)
		}
	})

	t.Run("env overrides default", func(t *testing.T) {
		opts := publishOptions{}
		fillPublishCredentials(&opts, func(key string) string {
			if key == "OSS_BASE_URL" {
				return "https://env.example.com/release"
			}
			return ""
		})
		if opts.baseURL != "https://env.example.com/release" {
			t.Fatalf("baseURL = %q, want env value", opts.baseURL)
		}
	})

	t.Run("default fallback", func(t *testing.T) {
		opts := publishOptions{}
		fillPublishCredentials(&opts, func(string) string { return "" })
		if opts.baseURL != release.DefaultBaseURL {
			t.Fatalf("baseURL = %q, want default %q", opts.baseURL, release.DefaultBaseURL)
		}
	})
}

func TestRunRejectsLegacyPublishInvocation(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"linux", "amd64", "2.4.0", "stable", "/tmp/erda-cli"}, func(string) string {
		return ""
	}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected legacy positional invocation to fail")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("expected unknown command error, got %q", err.Error())
	}
}

func TestRunAcceptsPublishAndPruneHelpSubcommands(t *testing.T) {
	testcases := []struct {
		name string
		args []string
	}{
		{name: "publish help", args: []string{"publish", "--help"}},
		{name: "prune help", args: []string{"prune", "--help"}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			err := run(tc.args, func(string) string { return "" }, &stdout, &stderr)
			if err != nil {
				t.Fatalf("expected help to succeed: %v", err)
			}
			if stdout.Len() == 0 {
				t.Fatal("expected help output")
			}
		})
	}
}
