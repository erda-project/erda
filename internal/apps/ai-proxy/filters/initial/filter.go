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

package initial

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "initial"
)

var (
	_ reverseproxy.RequestFilter  = (*Initial)(nil)
	_ reverseproxy.ResponseFilter = (*Initial)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type Initial struct {
	*reverseproxy.DefaultResponseFilter
}

// Enable is always true for initial filter
func (f *Initial) Enable(ctx context.Context, req *http.Request) bool {
	return true
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &Initial{DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter()}, nil
}

func (f *Initial) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	return reverseproxy.Continue, nil
}

func (f *Initial) OnResponseChunkImmutable(ctx context.Context, infor reverseproxy.HttpInfor, copiedChunk []byte) (signal reverseproxy.Signal, err error) {
	return reverseproxy.Continue, nil
}

func (f *Initial) OnResponseEOFImmutable(ctx context.Context, infor reverseproxy.HttpInfor, copiedChunk []byte) error {
	return nil
}
