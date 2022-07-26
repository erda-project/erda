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

package handlers

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/dop/search/pb"
)

type SearchType string

const (
	SearchTypeIssue SearchType = "issue"
)

const (
	PageSize = 10
)

func (s SearchType) String() string { return string(s) }

type Handler interface {
	BeginSearch(ctx context.Context, req *pb.SearchRequest)
	SetNexts(nexts ...Handler)
	GetResult() (*pb.SearchResponse, error)
}

type BaseSearch struct {
	sync.WaitGroup
	sync.Mutex
	Nexts    []Handler
	errors   []error
	contents []*pb.SearchResultContent
}

func (b *BaseSearch) SetNexts(handlers ...Handler) { b.Nexts = handlers }
func (b *BaseSearch) BeginSearch(ctx context.Context, req *pb.SearchRequest) {
	b.DoNexts(ctx, req)
}
func (b *BaseSearch) DoNexts(ctx context.Context, req *pb.SearchRequest) {
	for _, next := range b.Nexts {
		b.Add(1)
		go func() {
			defer b.Done()
			next.BeginSearch(ctx, req)
		}()
	}
	b.Wait()
}
func (b *BaseSearch) AppendError(err error) {
	b.Lock()
	defer b.Unlock()
	b.errors = append(b.errors, err)
}
func (b *BaseSearch) AppendContent(content *pb.SearchResultContent) {
	b.Lock()
	defer b.Unlock()
	if content != nil {
		b.contents = append(b.contents, content)
	}
}
func (b *BaseSearch) GetResult() (*pb.SearchResponse, error) {
	for _, next := range b.Nexts {
		contents, err := next.GetResult()
		if err != nil {
			b.AppendError(err)
		}
		for _, content := range contents.Data {
			b.AppendContent(content)
		}
	}
	res := &pb.SearchResponse{
		Data: b.contents,
	}
	if len(b.errors) > 0 {
		errMsg := make([]string, 0)
		for _, err := range b.errors {
			errMsg = append(errMsg, err.Error())
		}
		err := errors.New(strings.Join(errMsg, "\n"))
		return res, err
	}
	return res, nil
}

func NewBaseSearch(nexts ...Handler) *BaseSearch {
	return &BaseSearch{Nexts: nexts}
}
