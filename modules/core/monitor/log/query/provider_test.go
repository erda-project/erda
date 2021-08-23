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

package query

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

func mockProvider() *provider {
	return &provider{
		cqlQuery: &mockCqlQuery{},
	}
}

func gzipString(data string) []byte {
	d, err := gzipContent(data)
	if err != nil {
		log.Fatal(err)
	}
	return d
}

func gzipContent(content string) ([]byte, error) {
	reader, err := compressWithPipe(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

func compressWithPipe(reader io.Reader) (io.Reader, error) {
	pipeReader, pipeWriter := io.Pipe()
	gzipWriter := gzip.NewWriter(pipeWriter)

	var err error
	go func() {
		_, err = io.Copy(gzipWriter, reader)
		gzipWriter.Close()
		// subsequent reads from the read half of the pipe will
		// return no bytes and the error err, or EOF if err is nil.
		pipeWriter.CloseWithError(err)
	}()

	return pipeReader, err
}
