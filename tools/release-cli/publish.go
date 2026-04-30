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
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/cobra"

	"github.com/erda-project/erda/tools/cli/release"
)

func newPublishCmd(getenv func(string) string, stdout io.Writer) *cobra.Command {
	var opts publishOptions

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish all supported CLI release artifacts from a build directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPublishCredentials(&opts, getenv)
			if err := completePublishOptions(&opts); err != nil {
				return err
			}

			bucket, err := newBucket(opts.endpoint, opts.bucketName, opts.keyID, opts.keySecret)
			if err != nil {
				return err
			}
			return publishAll(bucket, opts, stdout)
		},
	}
	cmd.Flags().StringVar(&opts.version, "version", "", "release version to publish")
	cmd.Flags().StringVar(&opts.channel, "channel", "", "release channel, defaults to the channel detected from --version")
	cmd.Flags().StringVar(&opts.dir, "dir", "", "directory containing built CLI artifacts")
	cmd.Flags().StringVar(&opts.baseURL, "base-url", "", "public base URL for manifest artifact links (defaults to OSS_BASE_URL or release.DefaultBaseURL)")
	return cmd
}

func fillPublishCredentials(opts *publishOptions, getenv func(string) string) {
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	opts.keyID = strings.TrimSpace(getenv("ACCESS_KEY_ID"))
	opts.keySecret = strings.TrimSpace(getenv("ACCESS_KEY_SECRET"))
	opts.endpoint = firstNonEmpty(strings.TrimSpace(getenv("OSS_ENDPOINT")), defaultOSSEndpoint)
	opts.bucketName = firstNonEmpty(strings.TrimSpace(getenv("OSS_BUCKET_NAME")), defaultOSSBucketName)
	opts.baseURL = firstNonEmpty(strings.TrimSpace(opts.baseURL), strings.TrimSpace(getenv("OSS_BASE_URL")), release.DefaultBaseURL)
}

func completePublishOptions(opts *publishOptions) error {
	var missing []string
	if strings.TrimSpace(opts.keyID) == "" {
		missing = append(missing, "ACCESS_KEY_ID")
	}
	if strings.TrimSpace(opts.keySecret) == "" {
		missing = append(missing, "ACCESS_KEY_SECRET")
	}
	if strings.TrimSpace(opts.version) == "" {
		missing = append(missing, "version")
	}
	if strings.TrimSpace(opts.dir) == "" {
		missing = append(missing, "dir")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required argument(s): %s", strings.Join(missing, ", "))
	}

	detectedChannel, err := release.DetectChannel(opts.version)
	if err != nil {
		return err
	}
	if opts.channel == "" {
		opts.channel = detectedChannel
	}
	if err := release.ValidateChannel(opts.channel); err != nil {
		return err
	}
	if opts.channel != detectedChannel {
		return fmt.Errorf("channel %q does not match version %q (detected %q)", opts.channel, opts.version, detectedChannel)
	}
	return nil
}

func resolvePublishTargets(opts publishOptions) ([]publishTarget, error) {
	targets := make([]publishTarget, 0, len(releaseTargets))
	for _, target := range releaseTargets {
		file := filepath.Join(opts.dir, target.fileName)
		info, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("release artifact not found: %s", file)
			}
			return nil, err
		}
		if info.IsDir() {
			return nil, fmt.Errorf("release artifact is a directory: %s", file)
		}
		if err := validateArtifactVersion(file, opts.version); err != nil {
			return nil, err
		}

		targets = append(targets, publishTarget{
			keyID:      opts.keyID,
			keySecret:  opts.keySecret,
			endpoint:   opts.endpoint,
			bucketName: opts.bucketName,
			baseURL:    opts.baseURL,
			goos:       target.goos,
			goarch:     target.goarch,
			version:    opts.version,
			channel:    opts.channel,
			file:       file,
		})
	}
	return targets, nil
}

func validateArtifactVersion(filePath, expectedVersion string) error {
	payload, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if !bytes.Contains(payload, []byte(expectedVersion)) {
		return fmt.Errorf("release artifact %s does not contain expected version %q; rebuild binaries before publishing", filePath, expectedVersion)
	}
	return nil
}

func publishAll(bucket *oss.Bucket, opts publishOptions, stdout io.Writer) error {
	targets, err := resolvePublishTargets(opts)
	if err != nil {
		return err
	}
	for _, target := range targets {
		if err := publishTargetArtifact(bucket, target); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Published %s/%s from %s\n", target.goos, target.goarch, target.file)
	}
	return nil
}

func publishTargetArtifact(bucket *oss.Bucket, target publishTarget) error {
	objectACL := oss.ObjectACL(oss.ACLPublicRead)

	archivePath, cleanup, err := packageArtifact(target)
	if err != nil {
		return err
	}
	defer cleanup()

	manifest, err := buildManifest(target, archivePath)
	if err != nil {
		return err
	}

	artifactObject := release.ArtifactObjectName(target.goos, target.goarch, target.version)
	if err := bucket.PutObjectFromFile(artifactObject, archivePath, objectACL); err != nil {
		return err
	}
	if err := putManifest(bucket, release.VersionManifestObjectName(target.goos, target.goarch, target.version), manifest, objectACL); err != nil {
		return err
	}
	if err := putManifest(bucket, release.ChannelManifestObjectName(target.goos, target.goarch, target.channel), manifest, objectACL); err != nil {
		return err
	}

	index, err := getVersionIndex(bucket, target.goos, target.goarch, target.channel)
	if err != nil {
		return err
	}
	index = release.UpdateVersionIndex(index, *manifest, 10)
	return putVersionIndex(bucket, release.ChannelVersionsObjectName(target.goos, target.goarch, target.channel), index, objectACL)
}

func buildManifest(target publishTarget, artifactPath string) (*release.Manifest, error) {
	file, err := os.Open(artifactPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sum := sha256.New()
	if _, err := io.Copy(sum, file); err != nil {
		return nil, err
	}

	return &release.Manifest{
		Version:     target.version,
		Channel:     target.channel,
		OS:          target.goos,
		Arch:        target.goarch,
		URL:         strings.TrimRight(target.baseURL, "/") + "/" + release.ArtifactObjectName(target.goos, target.goarch, target.version),
		SHA256:      fmt.Sprintf("%x", sum.Sum(nil)),
		BuildTime:   os.Getenv("BUILD_TIME"),
		CommitID:    os.Getenv("COMMIT_ID"),
		PublishedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func putManifest(bucket *oss.Bucket, objectName string, manifest *release.Manifest, options ...oss.Option) error {
	payload, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return bucket.PutObject(objectName, strings.NewReader(string(payload)), options...)
}

func putVersionIndex(bucket *oss.Bucket, objectName string, index *release.VersionIndex, options ...oss.Option) error {
	payload, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return bucket.PutObject(objectName, strings.NewReader(string(payload)), options...)
}

func getVersionIndex(bucket *oss.Bucket, goos, goarch, channel string) (*release.VersionIndex, error) {
	objectName := release.ChannelVersionsObjectName(goos, goarch, channel)
	body, err := bucket.GetObject(objectName)
	if err != nil {
		var serviceErr oss.ServiceError
		if errors.As(err, &serviceErr) && serviceErr.StatusCode == 404 {
			return &release.VersionIndex{
				Channel: channel,
				OS:      goos,
				Arch:    goarch,
			}, nil
		}
		return nil, err
	}
	defer body.Close()

	var index release.VersionIndex
	if err := json.NewDecoder(body).Decode(&index); err != nil {
		return nil, err
	}
	return &index, nil
}
