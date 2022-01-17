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

package cleaner

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"gotest.tools/assert"

	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
)

// -go:generate mockgen -destination=./mock_loader_test.go -package cleaner -source=../loader/interface.go Interface
func Test_getSortedIndices_Should_Success(t *testing.T) {
	now := time.Now()

	ctrl := gomock.NewController(t)
	indices := NewMockInterface(ctrl)
	defer ctrl.Finish()

	indices.EXPECT().AllIndices().Return(&loader.IndexGroup{
		List: []*loader.IndexEntry{
			{
				Index:      "d",
				StoreBytes: 2,
				MaxT:       now.AddDate(0, 0, -3),
			},
			{
				Index:      "a",
				StoreBytes: 1,
				MaxT:       now.AddDate(0, 0, -1),
			},
			{
				Index:      "b",
				StoreBytes: 1,
				MaxT:       now.AddDate(0, 0, -2),
			},
			{
				Index:      "c",
				StoreBytes: 2,
				MaxT:       now.AddDate(0, 0, -2),
			},
		}})

	p := &provider{
		loader: indices,
	}

	want := []*loader.IndexEntry{
		{
			Index:      "d",
			StoreBytes: 2,
			MaxT:       now.AddDate(0, 0, -3),
		},
		{
			Index:      "b",
			StoreBytes: 1,
			MaxT:       now.AddDate(0, 0, -2),
		},
		{
			Index:      "c",
			StoreBytes: 2,
			MaxT:       now.AddDate(0, 0, -2),
		},
		{
			Index:      "a",
			StoreBytes: 1,
			MaxT:       now.AddDate(0, 0, -1),
		},
	}

	result := p.getSortedIndices()

	assert.DeepEqual(t, result, want)
}
