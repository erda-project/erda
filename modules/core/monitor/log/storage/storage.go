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

package storage

import (
	"context"
	"sort"

	"github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
)

type (
	// Operator .
	Operator int32
	// Filter .
	Filter struct {
		Key   string
		Op    Operator
		Value interface{}
	}

	// Selector .
	Selector struct {
		Start   int64
		End     int64
		Scheme  string
		Filters []*Filter
		Debug   bool
		Options map[string]interface{}
	}

	// Storage .
	Storage interface {
		NewWriter(ctx context.Context) (storekit.BatchWriter, error)
		Iterator(ctx context.Context, sel *Selector) (storekit.Iterator, error)
	}
)

const (
	// EQ equal
	EQ Operator = iota
	REGEXP
)

// Comparer .
type Comparer struct{}

// DefaultComparer
var DefaultComparer = Comparer{}

var _ storekit.Comparer = (*Comparer)(nil)

func (c Comparer) Compare(a, b interface{}) int {
	al, ok := a.(*pb.LogItem)
	if !ok {
		return -1
	}
	bl, ok := b.(*pb.LogItem)
	if !ok {
		return -1
	}
	return Compare(al, bl)
}

// Compare .
func Compare(al, bl *pb.LogItem) int {
	if al.Timestamp > bl.Timestamp {
		return 1
	} else if al.Timestamp < bl.Timestamp {
		return -1
	}
	if (al.Offset >= 0 && bl.Offset >= 0) || (al.Offset < 0 && bl.Offset < 0) {
		if al.Offset > bl.Offset {
			return 1
		} else if al.Offset < bl.Offset {
			return -1
		}
	} else if al.Offset < 0 {
		return 1
	} else if bl.Offset < 0 {
		return -1
	}
	return 0
}

// Logs .
type Logs []*pb.LogItem

var _ sort.Interface = (Logs)(nil)

func (l Logs) Len() int      { return len(l) }
func (l Logs) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l Logs) Less(i, j int) bool {
	return Compare(l[i], l[j]) < 0
}
