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
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
)

func BenchmarkGzipContentV1(b *testing.B) {
	s := strings.Builder{}
	for i := 0; i < 1000; i++ {
		s.WriteString("*")
	}
	b.ReportAllocs()

	content := s.String()
	for i := 0; i < b.N; i++ {
		gzipContent(content)
	}
}

func BenchmarkGzipContentV2(b *testing.B) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)

	s := strings.Builder{}
	for i := 0; i < 1000; i++ {
		s.WriteString("*")
	}
	b.ReportAllocs()

	content := s.String()
	for i := 0; i < b.N; i++ {
		gzipContentV2(content, w)
	}
}
