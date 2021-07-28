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
