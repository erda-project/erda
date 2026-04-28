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
	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/erda-project/erda/tools/cli/release"
)

const (
	defaultOSSEndpoint   = "oss-cn-hangzhou.aliyuncs.com"
	defaultOSSBucketName = "erda-release"
	artifactPrefix       = "cli"
)

var (
	defaultOSList   = []string{"darwin", "linux", "windows"}
	defaultArchList = []string{"amd64", "arm64"}
	releaseTargets  = []releaseTarget{
		{goos: "darwin", goarch: "arm64", fileName: "erda-cli"},
		{goos: "linux", goarch: "amd64", fileName: "erda-cli-linux"},
	}
)

type releaseTarget struct {
	goos     string
	goarch   string
	fileName string
}

type publishOptions struct {
	keyID      string
	keySecret  string
	endpoint   string
	bucketName string
	baseURL    string
	version    string
	channel    string
	dir        string
}

type publishTarget struct {
	keyID      string
	keySecret  string
	endpoint   string
	bucketName string
	baseURL    string
	goos       string
	goarch     string
	version    string
	channel    string
	file       string
}

type pruneOptions struct {
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
