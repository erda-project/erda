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

package table

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/tools/monitor/core/settings/retention-strategy"
)

func Test_formatTTLToDays(t *testing.T) {
	tests := []struct {
		ttl  *retention.TTL
		want int64
	}{
		{
			ttl: &retention.TTL{
				All: time.Hour * 7 * 24,
			},
			want: 7,
		},
		{
			ttl: &retention.TTL{
				All: time.Hour,
			},
			want: 1,
		},
		{
			ttl: &retention.TTL{
				All: time.Hour*8*24 + time.Hour,
			},
			want: 9,
		},
	}

	for _, tt := range tests {
		ret := FormatTTLToDays(tt.ttl)
		assert.Equal(t, tt.want, ret)
	}
}
