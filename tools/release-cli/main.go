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
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/cobra"

	"github.com/erda-project/erda/tools/cli/release"
)

const (
	defaultOSSEndpoint   = "oss-cn-hangzhou.aliyuncs.com"
	defaultOSSBucketName = "erda-release"
	artifactPrefix       = "cli"
)

var (
	defaultOSList           = []string{"darwin", "linux", "windows"}
	defaultArchList         = []string{"amd64", "arm64"}
	defaultReleaseArtifacts = []releaseArtifact{
		{goos: "darwin", goarch: "arm64", fileName: "erda-cli"},
		{goos: "linux", goarch: "amd64", fileName: "erda-cli-linux"},
	}
)

type releaseArtifact struct {
	goos     string
	goarch   string
	fileName string
}

type publishArgs struct {
	keyID      string
	keySecret  string
	endpoint   string
	bucketName string
	version    string
	channel    string
	dir        string
}

type releaseArgs struct {
	keyID      string
	keySecret  string
	endpoint   string
	bucketName string
	goos       string
	goarch     string
	version    string
	channel    string
	file       string
}

type pruneArgs struct {
	keyID      string
	keySecret  string
	endpoint   string
	bucketName string
	keep       int
	apply      bool
	channels   []string
}

type prunePlan struct {
	goos          string
	goarch        string
	channel       string
	retained      []release.Manifest
	deleteObjects []string
}

func main() {
	if err := run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}
}

func run(rawArgs []string, getenv func(string) string, stdout, stderr io.Writer) error {
	cmd := newRootCmd(getenv, stdout, stderr)
	cmd.SetArgs(rawArgs)
	return cmd.Execute()
}

func newRootCmd(getenv func(string) string, stdout, stderr io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "release-cli",
		Short:         "Publish and prune Erda CLI release artifacts",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Help(); err != nil {
				return err
			}
			return errors.New("missing subcommand")
		},
	}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.AddCommand(newPublishCmd(getenv, stdout), newPruneCmd(getenv, stdout))
	return rootCmd
}

func completePublishArgs(args *publishArgs) error {
	if strings.TrimSpace(args.keyID) == "" || strings.TrimSpace(args.keySecret) == "" {
		var missing []string
		if strings.TrimSpace(args.keyID) == "" {
			missing = append(missing, "ACCESS_KEY_ID")
		}
		if strings.TrimSpace(args.keySecret) == "" {
			missing = append(missing, "ACCESS_KEY_SECRET")
		}
		return fmt.Errorf("missing required argument(s): %s", strings.Join(missing, ", "))
	}
	if strings.TrimSpace(args.version) == "" {
		return errors.New("missing required argument(s): version")
	}
	if strings.TrimSpace(args.dir) == "" {
		return errors.New("missing required argument(s): dir")
	}
	detectedChannel, err := release.DetectChannel(args.version)
	if err != nil {
		return err
	}
	if args.channel == "" {
		args.channel = detectedChannel
	}
	if err := release.ValidateChannel(args.channel); err != nil {
		return err
	}
	if args.channel != detectedChannel {
		return fmt.Errorf("channel %q does not match version %q (detected %q)", args.channel, args.version, detectedChannel)
	}
	return nil
}

func resolvePublishReleases(args publishArgs) ([]releaseArgs, error) {
	var releases []releaseArgs
	for _, artifact := range defaultReleaseArtifacts {
		file := filepath.Join(args.dir, artifact.fileName)
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
		releases = append(releases, releaseArgs{
			keyID:      args.keyID,
			keySecret:  args.keySecret,
			endpoint:   args.endpoint,
			bucketName: args.bucketName,
			goos:       artifact.goos,
			goarch:     artifact.goarch,
			version:    args.version,
			channel:    args.channel,
			file:       file,
		})
	}
	return releases, nil
}

func parsePruneArgs(rawArgs []string, getenv func(string) string) (pruneArgs, error) {
	if getenv == nil {
		getenv = func(string) string { return "" }
	}

	args := pruneArgs{
		keyID:      strings.TrimSpace(getenv("ACCESS_KEY_ID")),
		keySecret:  strings.TrimSpace(getenv("ACCESS_KEY_SECRET")),
		endpoint:   firstNonEmpty(strings.TrimSpace(getenv("OSS_ENDPOINT")), defaultOSSEndpoint),
		bucketName: firstNonEmpty(strings.TrimSpace(getenv("OSS_BUCKET_NAME")), defaultOSSBucketName),
		keep:       10,
	}
	channelCSV := "alpha,beta"
	for i := 0; i < len(rawArgs); i++ {
		switch rawArgs[i] {
		case "--apply":
			args.apply = true
		case "--keep":
			if i+1 >= len(rawArgs) {
				return pruneArgs{}, errors.New("flag needs an argument: --keep")
			}
			i++
			var keep int
			if _, err := fmt.Sscanf(rawArgs[i], "%d", &keep); err != nil {
				return pruneArgs{}, fmt.Errorf("invalid value %q for --keep", rawArgs[i])
			}
			args.keep = keep
		case "--channel":
			if i+1 >= len(rawArgs) {
				return pruneArgs{}, errors.New("flag needs an argument: --channel")
			}
			i++
			channelCSV = rawArgs[i]
		default:
			return pruneArgs{}, fmt.Errorf("unexpected argument(s): %s", strings.Join(rawArgs[i:], ", "))
		}
	}
	return args, completePruneArgs(&args, channelCSV)
}

func completePruneArgs(args *pruneArgs, channelCSV string) error {
	if args.keyID == "" || args.keySecret == "" {
		return errors.New("ACCESS_KEY_ID and ACCESS_KEY_SECRET are required")
	}
	if args.keep < 1 {
		return errors.New("--keep must be a positive integer")
	}

	args.channels = args.channels[:0]
	for _, channel := range strings.Split(channelCSV, ",") {
		channel = strings.TrimSpace(channel)
		if channel == "" {
			continue
		}
		if err := release.ValidateChannel(channel); err != nil {
			return err
		}
		args.channels = append(args.channels, channel)
	}
	if len(args.channels) == 0 {
		return errors.New("at least one channel is required")
	}

	return nil
}

func newPublishCmd(getenv func(string) string, stdout io.Writer) *cobra.Command {
	var args publishArgs

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish all supported CLI release artifacts from a build directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, rawArgs []string) error {
			if getenv == nil {
				getenv = func(string) string { return "" }
			}
			args.keyID = strings.TrimSpace(getenv("ACCESS_KEY_ID"))
			args.keySecret = strings.TrimSpace(getenv("ACCESS_KEY_SECRET"))
			args.endpoint = firstNonEmpty(strings.TrimSpace(getenv("OSS_ENDPOINT")), defaultOSSEndpoint)
			args.bucketName = firstNonEmpty(strings.TrimSpace(getenv("OSS_BUCKET_NAME")), defaultOSSBucketName)
			if err := completePublishArgs(&args); err != nil {
				return err
			}
			bucket, err := newBucket(args.endpoint, args.bucketName, args.keyID, args.keySecret)
			if err != nil {
				return err
			}
			if err := publishReleases(bucket, args, stdout); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&args.version, "version", "", "release version to publish")
	cmd.Flags().StringVar(&args.channel, "channel", "", "release channel, defaults to the channel detected from --version")
	cmd.Flags().StringVar(&args.dir, "dir", "", "directory containing built CLI artifacts")
	return cmd
}

func newPruneCmd(getenv func(string) string, stdout io.Writer) *cobra.Command {
	var args pruneArgs
	var channelCSV string

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune old CLI release artifacts and rebuild version indexes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, rawArgs []string) error {
			if getenv == nil {
				getenv = func(string) string { return "" }
			}
			args.keyID = strings.TrimSpace(getenv("ACCESS_KEY_ID"))
			args.keySecret = strings.TrimSpace(getenv("ACCESS_KEY_SECRET"))
			args.endpoint = firstNonEmpty(strings.TrimSpace(getenv("OSS_ENDPOINT")), defaultOSSEndpoint)
			args.bucketName = firstNonEmpty(strings.TrimSpace(getenv("OSS_BUCKET_NAME")), defaultOSSBucketName)
			if err := completePruneArgs(&args, channelCSV); err != nil {
				return err
			}
			bucket, err := newBucket(args.endpoint, args.bucketName, args.keyID, args.keySecret)
			if err != nil {
				return err
			}
			return pruneReleases(bucket, args, stdout)
		},
	}
	cmd.Flags().IntVar(&args.keep, "keep", 10, "keep most recent N versions per os/arch/channel")
	cmd.Flags().BoolVar(&args.apply, "apply", false, "apply deletion and index update")
	cmd.Flags().StringVar(&channelCSV, "channel", "alpha,beta", "comma-separated channels to prune")
	return cmd
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func newBucket(endpoint, bucketName, keyID, keySecret string) (*oss.Bucket, error) {
	client, err := oss.New(endpoint, keyID, keySecret)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketName)
}

func publishRelease(bucket *oss.Bucket, args releaseArgs) error {
	objectAcl := oss.ObjectACL(oss.ACLPublicRead)
	archivePath, cleanup, err := packageArtifact(args)
	if err != nil {
		return err
	}
	defer cleanup()

	manifest, err := buildManifest(args, archivePath)
	if err != nil {
		return err
	}

	artifactObjectName := release.ArtifactObjectName(args.goos, args.goarch, args.version)
	if err = bucket.PutObjectFromFile(artifactObjectName, archivePath, objectAcl); err != nil {
		return err
	}
	if err = putManifest(bucket, release.VersionManifestObjectName(args.goos, args.goarch, args.version), manifest, objectAcl); err != nil {
		return err
	}
	if err = putManifest(bucket, release.ChannelManifestObjectName(args.goos, args.goarch, args.channel), manifest, objectAcl); err != nil {
		return err
	}

	index, err := getVersionIndex(bucket, args.goos, args.goarch, args.channel)
	if err != nil {
		return err
	}
	index = release.UpdateVersionIndex(index, *manifest, 10)
	if err = putVersionIndex(bucket, release.ChannelVersionsObjectName(args.goos, args.goarch, args.channel), index, objectAcl); err != nil {
		return err
	}

	return nil
}

func publishReleases(bucket *oss.Bucket, args publishArgs, stdout io.Writer) error {
	releases, err := resolvePublishReleases(args)
	if err != nil {
		return err
	}
	for _, releaseArgs := range releases {
		if err := publishRelease(bucket, releaseArgs); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Published %s/%s from %s\n", releaseArgs.goos, releaseArgs.goarch, releaseArgs.file)
	}
	return nil
}

func pruneReleases(bucket *oss.Bucket, args pruneArgs, stdout io.Writer) error {
	var plans []prunePlan
	for _, goos := range defaultOSList {
		for _, goarch := range defaultArchList {
			for _, channel := range args.channels {
				plan, ok, err := buildPrunePlan(bucket, goos, goarch, channel, args.keep)
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
	totalIndexes := 0
	fmt.Fprintf(stdout, "Mode: %s\n", map[bool]string{true: "apply", false: "dry-run"}[args.apply])
	fmt.Fprintf(stdout, "Bucket: %s\n", args.bucketName)
	fmt.Fprintf(stdout, "Endpoint: %s\n", args.endpoint)
	fmt.Fprintf(stdout, "Channels: %s\n", strings.Join(args.channels, ","))
	fmt.Fprintf(stdout, "Keep count: %d\n\n", args.keep)

	for _, plan := range plans {
		totalDeletes += len(plan.deleteObjects)
		totalIndexes++
		if len(plan.deleteObjects) == 0 {
			fmt.Fprintf(stdout, "[KEEP] %s/%s/%s: total=%d, keep=%d, delete=0\n", plan.goos, plan.goarch, plan.channel, len(plan.retained), args.keep)
		} else {
			fmt.Fprintf(stdout, "[PLAN] %s/%s/%s: total=%d, keep=%d, delete=%d\n", plan.goos, plan.goarch, plan.channel, len(plan.retained)+len(plan.deleteObjects)/2, args.keep, len(plan.deleteObjects)/2)
			for _, objectName := range plan.deleteObjects {
				fmt.Fprintf(stdout, "  - oss://%s/%s\n", args.bucketName, objectName)
			}
		}
		fmt.Fprintf(stdout, "  - rebuild oss://%s/%s with %d version(s)\n", args.bucketName, release.ChannelVersionsObjectName(plan.goos, plan.goarch, plan.channel), len(plan.retained))
	}

	fmt.Fprintf(stdout, "\nPlanned delete objects: %d\n", totalDeletes)
	fmt.Fprintf(stdout, "Planned index updates: %d\n", totalIndexes)
	if !args.apply {
		fmt.Fprintln(stdout, "Dry-run only. Re-run with --apply to delete objects and upload rebuilt version indexes.")
		return nil
	}

	objectAcl := oss.ObjectACL(oss.ACLPublicRead)
	for _, plan := range plans {
		for _, objectName := range plan.deleteObjects {
			fmt.Fprintf(stdout, "[DELETE] oss://%s/%s\n", args.bucketName, objectName)
			if err := bucket.DeleteObject(objectName); err != nil {
				return err
			}
		}
		index := &release.VersionIndex{
			Channel:  plan.channel,
			OS:       plan.goos,
			Arch:     plan.goarch,
			Versions: plan.retained,
		}
		indexObjectName := release.ChannelVersionsObjectName(plan.goos, plan.goarch, plan.channel)
		fmt.Fprintf(stdout, "[UPLOAD] oss://%s/%s (%d version(s))\n", args.bucketName, indexObjectName, len(plan.retained))
		if err := putVersionIndex(bucket, indexObjectName, index, objectAcl); err != nil {
			return err
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
	if len(plan.retained) > keep {
		for _, manifest := range plan.retained[keep:] {
			plan.deleteObjects = append(plan.deleteObjects,
				release.ArtifactObjectName(manifest.OS, manifest.Arch, manifest.Version),
				release.VersionManifestObjectName(manifest.OS, manifest.Arch, manifest.Version),
			)
		}
		plan.retained = plan.retained[:keep]
	}

	return plan, true, nil
}

func listChannelManifests(bucket *oss.Bucket, goos, goarch, channel string) ([]release.Manifest, error) {
	prefix := path.Join(artifactPrefix, goos, goarch) + "/"
	var manifests []release.Manifest
	marker := ""

	for {
		result, err := bucket.ListObjects(
			oss.Prefix(prefix),
			oss.Marker(marker),
			oss.MaxKeys(1000),
		)
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
			if manifest.Channel != channel {
				continue
			}
			manifests = append(manifests, *manifest)
		}

		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
	}

	return manifests, nil
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

func buildManifest(args releaseArgs, artifactPath string) (*release.Manifest, error) {
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
		Version:     args.version,
		Channel:     args.channel,
		OS:          args.goos,
		Arch:        args.goarch,
		URL:         release.DefaultBaseURL + "/" + release.ArtifactObjectName(args.goos, args.goarch, args.version),
		SHA256:      fmt.Sprintf("%x", sum.Sum(nil)),
		BuildTime:   os.Getenv("BUILD_TIME"),
		CommitID:    os.Getenv("COMMIT_ID"),
		PublishedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func packageArtifact(args releaseArgs) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "erda-cli-release-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	archivePath := filepath.Join(tmpDir, filepath.Base(release.ArtifactObjectName(args.goos, args.goarch, args.version)))
	switch args.goos {
	case "windows":
		err = packageZipArchive(archivePath, release.ExecutableFileName(args.goos), args.file)
	default:
		err = packageTarGzArchive(archivePath, release.ExecutableFileName(args.goos), args.file)
	}
	if err != nil {
		cleanup()
		return "", nil, err
	}
	return archivePath, cleanup, nil
}

func packageTarGzArchive(archivePath, entryName, sourcePath string) (err error) {
	src, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := dst.Close(); err == nil {
			err = closeErr
		}
	}()

	gw := gzip.NewWriter(dst)
	defer func() {
		if closeErr := gw.Close(); err == nil {
			err = closeErr
		}
	}()

	tw := tar.NewWriter(gw)
	defer func() {
		if closeErr := tw.Close(); err == nil {
			err = closeErr
		}
	}()

	header := &tar.Header{
		Name: entryName,
		Mode: int64(info.Mode().Perm()),
		Size: info.Size(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err := io.Copy(tw, src); err != nil {
		return err
	}
	return nil
}

func packageZipArchive(archivePath, entryName, sourcePath string) (err error) {
	src, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := dst.Close(); err == nil {
			err = closeErr
		}
	}()

	zw := zip.NewWriter(dst)
	defer func() {
		if closeErr := zw.Close(); err == nil {
			err = closeErr
		}
	}()

	w, err := zw.Create(entryName)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, src); err != nil {
		return err
	}
	return nil
}

func putManifest(bucket *oss.Bucket, objectName string, manifest *release.Manifest, objectAcl oss.Option) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return bucket.PutObject(objectName, bytes.NewReader(data), objectAcl)
}

func putVersionIndex(bucket *oss.Bucket, objectName string, index *release.VersionIndex, objectAcl oss.Option) error {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return bucket.PutObject(objectName, bytes.NewReader(data), objectAcl)
}

func getVersionIndex(bucket *oss.Bucket, goos, goarch, channel string) (*release.VersionIndex, error) {
	objectName := release.ChannelVersionsObjectName(goos, goarch, channel)
	body, err := bucket.GetObject(objectName)
	if err != nil {
		var serviceErr oss.ServiceError
		if errors.As(err, &serviceErr) && serviceErr.StatusCode == 404 {
			return &release.VersionIndex{Channel: channel, OS: goos, Arch: goarch}, nil
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
