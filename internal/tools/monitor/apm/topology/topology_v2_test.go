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

package topology

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
)

func Test_queryConditionsV2(t *testing.T) {
	type args struct {
		params  Vo
		tagInfo *TagInfo
	}
	tests := []struct {
		name    string
		args    args
		wantsql string
	}{
		{
			name: "without ApplicationName",
			args: args{
				params: Vo{
					TerminusKey: "123",
					StartTime:   1660192621035,
					EndTime:     1660196221035,
				},
				tagInfo: &TagInfo{},
			},
			wantsql: `SELECT * FROM "table" WHERE (("tenant_id" IN ('123', '')) AND ("timestamp" >= fromUnixTimestamp64Milli(toInt64(1660192621035))) AND ("timestamp" <= fromUnixTimestamp64Milli(toInt64(1660196221035))) AND ((tag_values[indexOf(tag_keys, 'target_terminus_key')] = '123') OR (tag_values[indexOf(tag_keys, 'source_terminus_key')] = '123')) AND (tag_values[indexOf(tag_keys, 'component')] NOT IN ('registerCenter', 'configCenter', 'noticeCenter')) AND (tag_values[indexOf(tag_keys, 'target_addon_type')] NOT IN ('registerCenter', 'configCenter', 'noticeCenter')))`,
		},
		{
			name: "with ApplicationName",
			args: args{
				params: Vo{
					TerminusKey: "123",
					StartTime:   1660192621035,
					EndTime:     1660196221035,
				},
				tagInfo: &TagInfo{
					ApplicationName: "hello",
				},
			},
			wantsql: `SELECT * FROM "table" WHERE (("tenant_id" IN ('123', '')) AND ("timestamp" >= fromUnixTimestamp64Milli(toInt64(1660192621035))) AND ("timestamp" <= fromUnixTimestamp64Milli(toInt64(1660196221035))) AND ((tag_values[indexOf(tag_keys, 'target_terminus_key')] = '123') OR (tag_values[indexOf(tag_keys, 'source_terminus_key')] = '123')) AND (tag_values[indexOf(tag_keys, 'component')] NOT IN ('registerCenter', 'configCenter', 'noticeCenter')) AND (tag_values[indexOf(tag_keys, 'target_addon_type')] NOT IN ('registerCenter', 'configCenter', 'noticeCenter')) AND ((tag_values[indexOf(tag_keys, 'application_name')] = 'hello') OR (tag_values[indexOf(tag_keys, 'target_application_name')] = 'hello') OR (tag_values[indexOf(tag_keys, 'source_application_name')] = 'hello')))`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd := queryConditionsV2(goqu.From("table"), tt.args.params, tt.args.tagInfo, HttpRecMircoIndexType)
			sqlstr, _, err := sd.ToSQL()
			if err != nil {
				assert.Fail(t, "must not error: %s", err)
				return
			}
			assert.Equalf(t, tt.wantsql, sqlstr, "queryConditionsV2(goqu.From(), %v, %v)", tt.args.params, tt.args.tagInfo)
		})
	}
}

func Test_graphTopo_toNodes(t *testing.T) {
	g := &graphTopo{
		adj:   map[string]map[string]struct{}{},
		nodes: map[string]Node{},
	}
	stats := Metric{Count: 1}

	g.addDirectEdge(mockRelation("a", "b", stats))
	g.addDirectEdge(mockRelation("a", "c", stats))
	g.addDirectEdge(mockRelation("b", "a", stats))
	g.addDirectEdge(mockRelation("b", "c", stats))
	g.addDirectEdge(mockRelation("c", "a", stats))
	g.addDirectEdge(mockRelation("c", "b", stats))

	nodes := g.toNodes()
	for _, n := range nodes {
		assert.Equal(t, int64(4), n.Metric.Count)
		assert.Equal(t, 2, len(n.Parents))
	}
}

func mockRelation(sid, tid string, stats Metric) *relation {
	return &relation{source: Node{Id: sid}, target: Node{Id: tid}, stats: stats}
}

func Test_handleTargetOtherNodesByHttpUrl(t *testing.T) {
	opts := targetOtherNodeOptions{MaxNum: 3}

	// count <= 3, return directly
	allRels := []*relation{
		{
			source: Node{Type: TypeService, Name: "monitor"},
			target: Node{Type: TypeExternal, Name: "https://erda.cloud"},
		},
	}
	handleTargetOtherNodesByHttpUrl(allRels, opts)
	assert.Equal(t, allRels[0].target.Name, "https://erda.cloud")

	// count = 4, trigger remove query param, and got 1
	allRels = []*relation{
		{
			source: Node{Type: TypeService, Name: "monitor"},
			target: Node{Type: TypeExternal, Name: "https://erda.cloud"},
		},
		{
			source: Node{Type: TypeService, Name: "monitor"},
			target: Node{Type: TypeExternal, Name: "https://erda.cloud?a=b"},
		},
		{
			source: Node{Type: TypeService, Name: "monitor"},
			target: Node{Type: TypeExternal, Name: "https://erda.cloud?a=c"},
		},
		{
			source: Node{Type: TypeService, Name: "monitor"},
			target: Node{Type: TypeExternal, Name: "https://erda.cloud?a=d"},
		},
	}
	handleTargetOtherNodesByHttpUrl(allRels, opts)
	newNodeId := calculateTargetOtherNodeId(&Node{Type: TypeExternal, Name: "https://erda.cloud"})
	for _, rel := range allRels {
		assert.Equal(t, newNodeId, rel.target.Id)
	}

	// count = 4, trigger remove query param, and got 4, so trigger only host
	allRels = []*relation{
		{
			source: Node{Type: TypeService, Name: "monitor"},
			target: Node{Type: TypeExternal, Name: "https://erda.cloud"},
		},
		{
			source: Node{Type: TypeService, Name: "monitor"},
			target: Node{Type: TypeExternal, Name: "http://erda.cloud?a=b"},
		},
		{
			source: Node{Type: TypeService, Name: "monitor"},
			target: Node{Type: TypeExternal, Name: "https://www.erda.cloud?a=c"},
		},
		{
			source: Node{Type: TypeService, Name: "monitor"},
			target: Node{Type: TypeExternal, Name: "https://github.com/erda-project/erda?a=d"},
		},
	}
	handleTargetOtherNodesByHttpUrl(allRels, opts)
	assert.Equal(t, allRels[0].target.Name, "erda.cloud")
	assert.Equal(t, calculateTargetOtherNodeId(&Node{Name: "erda.cloud", Type: TypeExternal}), allRels[0].target.Id)
	assert.Equal(t, allRels[1].target.Name, "erda.cloud")
	assert.Equal(t, calculateTargetOtherNodeId(&Node{Name: "erda.cloud", Type: TypeExternal}), allRels[1].target.Id)
	assert.Equal(t, allRels[2].target.Name, "www.erda.cloud")
	assert.Equal(t, calculateTargetOtherNodeId(&Node{Name: "www.erda.cloud", Type: TypeExternal}), allRels[2].target.Id)
	assert.Equal(t, allRels[3].target.Name, "github.com")
	assert.Equal(t, calculateTargetOtherNodeId(&Node{Name: "github.com", Type: TypeExternal}), allRels[3].target.Id)
}
