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

package compressor

import (
	"compress/gzip"
	"io"
	"sync"

	"github.com/golang/snappy"
)

func GetGzipReader(r io.Reader) (*gzip.Reader, error) {
	v := gzipReaderPool.Get()
	if v == nil {
		return gzip.NewReader(r)
	}
	zr := v.(*gzip.Reader)
	if err := zr.Reset(r); err != nil {
		return nil, err
	}
	return zr, nil
}

func PutGzipReader(zr *gzip.Reader) {
	_ = zr.Close()
	gzipReaderPool.Put(zr)
}

// Reused gzip reader
var gzipReaderPool sync.Pool

func GetSnappyReader(r io.Reader) *snappy.Reader {
	v := snappyReaderPoll.Get()
	if v == nil {
		return snappy.NewReader(r)
	}
	zr := v.(*snappy.Reader)
	zr.Reset(r)
	return zr
}

func PutSnappyReader(zr *snappy.Reader) {
	snappyReaderPoll.Put(zr)
}

var snappyReaderPoll sync.Pool
