// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"io"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OSS struct {
	endpoint        string
	accessKeyID     string
	accessKeySecret string
	bucket          string
	clientOptions   []oss.ClientOption
	options         []oss.Option
}

func NewOSS(endpoint, accessKeyID, accessKeySecret, bucket string,
	clientOptions []oss.ClientOption, options []oss.Option) *OSS {
	var o OSS
	o.endpoint = endpoint
	o.accessKeyID = accessKeyID
	o.accessKeySecret = accessKeySecret
	o.bucket = bucket
	o.clientOptions = clientOptions
	o.options = options
	return &o
}

func (o *OSS) Type() Type {
	return TypeOSS
}

func (o *OSS) Read(path string) (io.Reader, error) {
	path = handlePath(path)
	client, err := o.newClient()
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(o.bucket)
	if err != nil {
		return nil, err
	}
	return bucket.GetObject(path, o.options...)
}

func (o *OSS) Write(path string, r io.Reader) error {
	path = handlePath(path)
	client, err := o.newClient()
	if err != nil {
		return err
	}
	bucket, err := client.Bucket(o.bucket)
	if err != nil {
		return err
	}
	return bucket.PutObject(path, r, o.options...)
}

func (o *OSS) Delete(path string) error {
	path = handlePath(path)
	client, err := o.newClient()
	if err != nil {
		return err
	}
	bucket, err := client.Bucket(o.bucket)
	if err != nil {
		return err
	}
	return bucket.DeleteObject(path)
}

func (o *OSS) newClient() (*oss.Client, error) {
	return oss.New(o.endpoint, o.accessKeyID, o.accessKeySecret, o.clientOptions...)
}

// handlePath
// path cannot start with "/" or "\", see: vendor/github.com/aliyun/aliyun-oss-go-sdk/oss/bucket.go:28
func handlePath(path string) string {
	return strings.TrimPrefix(path, "/")
}
