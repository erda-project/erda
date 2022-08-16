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

package clickhousesource

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
)

func Test_mergeNodeType(t *testing.T) {
	type args struct {
		target *NodeType
		source *NodeType
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			args: args{
				target: TargetServiceNodeType,
				source: SourceServiceNodeType,
			},
			want: 1732,
		},
		{
			args: args{
				target: nil,
				source: nil,
			},
			want: 32,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := mergeNodeType(tt.args.target, tt.args.source)
			sd := goqu.From("table").Select(agg.Select()).Where(agg.Where()).GroupBy(agg.GroupBy())
			sqlstr, _, err := sd.ToSQL()
			if err != nil {
				assert.Fail(t, "must not error: %s", err)
				return
			}
			assert.Equal(t, tt.want, len(sqlstr))
		})
	}
}
