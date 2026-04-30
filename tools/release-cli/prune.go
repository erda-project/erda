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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/cobra"

	"github.com/erda-project/erda/tools/cli/release"
)

func newPruneCmd(getenv func(string) string, stdout io.Writer) *cobra.Command {
	var opts pruneOptions
	var channelCSV string

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune old CLI release artifacts and rebuild version indexes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPruneCredentials(&opts, getenv)
			if err := completePruneOptions(&opts, channelCSV); err != nil {
				return err
			}

			bucket, err := newBucket(opts.endpoint, opts.bucketName, opts.keyID, opts.keySecret)
			if err != nil {
				return err
			}
			return pruneReleases(bucket, opts, stdout)
		},
	}
	cmd.Flags().IntVar(&opts.keep, "keep", 10, "keep most recent N versions per os/arch/channel")
	cmd.Flags().BoolVar(&opts.apply, "apply", false, "apply deletion and index update")
	cmd.Flags().StringVar(&channelCSV, "channel", "alpha,beta", "comma-separated channels to prune")
	return cmd
}

func fillPruneCredentials(opts *pruneOptions, getenv func(string) string) {
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	opts.keyID = strings.TrimSpace(getenv("ACCESS_KEY_ID"))
	opts.keySecret = strings.TrimSpace(getenv("ACCESS_KEY_SECRET"))
	opts.endpoint = firstNonEmpty(strings.TrimSpace(getenv("OSS_ENDPOINT")), defaultOSSEndpoint)
	opts.bucketName = firstNonEmpty(strings.TrimSpace(getenv("OSS_BUCKET_NAME")), defaultOSSBucketName)
}

func completePruneOptions(opts *pruneOptions, channelCSV string) error {
	if opts.keyID == "" || opts.keySecret == "" {
		return errors.New("ACCESS_KEY_ID and ACCESS_KEY_SECRET are required")
	}
	if opts.keep < 1 {
		return errors.New("--keep must be a positive integer")
	}

	opts.channels = opts.channels[:0]
	for _, channel := range strings.Split(channelCSV, ",") {
		channel = strings.TrimSpace(channel)
		if channel == "" {
			continue
		}
		if err := release.ValidateChannel(channel); err != nil {
			return err
		}
		opts.channels = append(opts.channels, channel)
	}
	if len(opts.channels) == 0 {
		return errors.New("at least one channel is required")
	}
	return nil
}

func pruneReleases(bucket *oss.Bucket, opts pruneOptions, stdout io.Writer) error {
	var plans []prunePlan
	for _, goos := range defaultOSList {
		for _, goarch := range defaultArchList {
			for _, channel := range opts.channels {
				plan, ok, err := buildPrunePlan(bucket, goos, goarch, channel, opts.keep)
				if err != nil {
					return err
				}
				if ok {
					plans = append(plans, plan)
				}
			}
		}
	}

	totalDeletes := 0
	totalManifestUpdates := 0
	totalIndexes := 0
	fmt.Fprintf(stdout, "Mode: %s\n", map[bool]string{true: "apply", false: "dry-run"}[opts.apply])
	fmt.Fprintf(stdout, "Bucket: %s\n", opts.bucketName)
	fmt.Fprintf(stdout, "Endpoint: %s\n", opts.endpoint)
	fmt.Fprintf(stdout, "Channels: %s\n", strings.Join(opts.channels, ","))
	fmt.Fprintf(stdout, "Keep count: %d\n\n", opts.keep)

	for _, plan := range plans {
		totalDeletes += len(plan.deleteObjects)
		totalManifestUpdates++
		totalIndexes++
		if len(plan.deleteObjects) == 0 {
			fmt.Fprintf(stdout, "[KEEP] %s/%s/%s: total=%d, keep=%d, delete=0\n", plan.goos, plan.goarch, plan.channel, len(plan.retained), opts.keep)
		} else {
			fmt.Fprintf(stdout, "[PLAN] %s/%s/%s: total=%d, keep=%d, delete=%d\n", plan.goos, plan.goarch, plan.channel, len(plan.retained)+len(plan.deleteObjects)/2, opts.keep, len(plan.deleteObjects)/2)
			for _, objectName := range plan.deleteObjects {
				fmt.Fprintf(stdout, "  - oss://%s/%s\n", opts.bucketName, objectName)
			}
		}
		fmt.Fprintf(stdout, "  - update oss://%s/%s to version %s\n", opts.bucketName, release.ChannelManifestObjectName(plan.goos, plan.goarch, plan.channel), plan.retained[0].Version)
		fmt.Fprintf(stdout, "  - rebuild oss://%s/%s with %d version(s)\n", opts.bucketName, release.ChannelVersionsObjectName(plan.goos, plan.goarch, plan.channel), len(plan.retained))
	}

	fmt.Fprintf(stdout, "\nPlanned delete objects: %d\n", totalDeletes)
	fmt.Fprintf(stdout, "Planned channel manifest updates: %d\n", totalManifestUpdates)
	fmt.Fprintf(stdout, "Planned index updates: %d\n", totalIndexes)
	if !opts.apply {
		fmt.Fprintln(stdout, "Dry-run only. Re-run with --apply to update channel manifests, upload rebuilt version indexes, and delete old objects.")
		return nil
	}

	objectACL := oss.ObjectACL(oss.ACLPublicRead)
	for _, plan := range plans {
		channelManifestObjectName := release.ChannelManifestObjectName(plan.goos, plan.goarch, plan.channel)
		channelManifest := plan.retained[0]
		fmt.Fprintf(stdout, "[UPLOAD] oss://%s/%s (version %s)\n", opts.bucketName, channelManifestObjectName, channelManifest.Version)
		if err := putManifest(bucket, channelManifestObjectName, &channelManifest, objectACL); err != nil {
			return err
		}

		index := &release.VersionIndex{
			Channel:  plan.channel,
			OS:       plan.goos,
			Arch:     plan.goarch,
			Versions: plan.retained,
		}
		indexObjectName := release.ChannelVersionsObjectName(plan.goos, plan.goarch, plan.channel)
		fmt.Fprintf(stdout, "[UPLOAD] oss://%s/%s (%d version(s))\n", opts.bucketName, indexObjectName, len(plan.retained))
		if err := putVersionIndex(bucket, indexObjectName, index, objectACL); err != nil {
			return err
		}
		for _, objectName := range plan.deleteObjects {
			fmt.Fprintf(stdout, "[DELETE] oss://%s/%s\n", opts.bucketName, objectName)
			if err := bucket.DeleteObject(objectName); err != nil {
				return err
			}
		}
	}

	fmt.Fprintln(stdout, "Done.")
	return nil
}

func buildPrunePlan(bucket *oss.Bucket, goos, goarch, channel string, keep int) (prunePlan, bool, error) {
	manifests, err := listChannelManifests(bucket, goos, goarch, channel)
	if err != nil {
		return prunePlan{}, false, err
	}
	if len(manifests) == 0 {
		return prunePlan{}, false, nil
	}

	index := &release.VersionIndex{Channel: channel, OS: goos, Arch: goarch}
	for _, manifest := range manifests {
		index = release.UpdateVersionIndex(index, manifest, 0)
	}

	plan := prunePlan{
		goos:     goos,
		goarch:   goarch,
		channel:  channel,
		retained: index.Versions,
	}
	if len(plan.retained) <= keep {
		return plan, true, nil
	}

	for _, manifest := range plan.retained[keep:] {
		plan.deleteObjects = append(plan.deleteObjects,
			release.ArtifactObjectName(manifest.OS, manifest.Arch, manifest.Version),
			release.VersionManifestObjectName(manifest.OS, manifest.Arch, manifest.Version),
		)
	}
	plan.retained = plan.retained[:keep]
	return plan, true, nil
}

func listChannelManifests(bucket *oss.Bucket, goos, goarch, channel string) ([]release.Manifest, error) {
	prefix := path.Join(artifactPrefix, goos, goarch) + "/"
	var (
		manifests []release.Manifest
		marker    string
	)

	for {
		result, err := bucket.ListObjects(oss.Prefix(prefix), oss.Marker(marker), oss.MaxKeys(1000))
		if err != nil {
			return nil, err
		}

		for _, object := range result.Objects {
			if !isVersionManifestObject(path.Base(object.Key), channel) {
				continue
			}
			manifest, err := fetchManifestObject(bucket, object.Key)
			if err != nil {
				return nil, err
			}
			if manifest.Channel == channel {
				manifests = append(manifests, *manifest)
			}
		}

		if !result.IsTruncated {
			return manifests, nil
		}
		marker = result.NextMarker
	}
}

func isVersionManifestObject(baseName, channel string) bool {
	if !strings.HasPrefix(baseName, "erda-cli-") || !strings.HasSuffix(baseName, ".json") {
		return false
	}
	if baseName == channel+".json" || baseName == channel+"-versions.json" {
		return false
	}
	version := strings.TrimSuffix(strings.TrimPrefix(baseName, "erda-cli-"), ".json")
	detectedChannel, err := release.DetectChannel(version)
	return err == nil && detectedChannel == channel
}

func fetchManifestObject(bucket *oss.Bucket, objectName string) (*release.Manifest, error) {
	body, err := bucket.GetObject(objectName)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var manifest release.Manifest
	if err := json.NewDecoder(body).Decode(&manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}
