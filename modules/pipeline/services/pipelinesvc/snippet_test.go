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

package pipelinesvc

import (
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/pipeline_snippet_client"
)

func TestHandleQueryPipelineYamlBySnippetConfigs(t *testing.T) {

	var svc = &PipelineSvc{}
	guard := monkey.Patch(pipeline_snippet_client.BatchGetSnippetPipelineYml, func(snippetConfig []apistructs.SnippetConfig) ([]apistructs.BatchSnippetConfigYml, error) {
		var result []apistructs.BatchSnippetConfigYml
		for _, v := range snippetConfig {
			result = append(result, apistructs.BatchSnippetConfigYml{
				Config: v,
				Yml:    v.ToString(),
			})
		}
		return result, nil
	})
	guard1 := monkey.Patch(pipeline_snippet_client.GetSnippetPipelineYml, func(snippetConfig apistructs.SnippetConfig) (string, error) {
		return snippetConfig.ToString(), nil
	})
	defer guard.Unpatch()
	defer guard1.Unpatch()

	var table = []struct {
		sourceSnippetConfigs []apistructs.SnippetConfig
	}{
		{
			sourceSnippetConfigs: []apistructs.SnippetConfig{
				{
					Source: "autotest",
					Name:   "custom",
					Labels: map[string]string{
						"key3": "key",
					},
				},
				{
					Source: "local",
					Name:   "pipeline",
					Labels: map[string]string{
						"key3": "key",
					},
				},
				{
					Source: "autotest",
					Name:   "custom",
					Labels: map[string]string{
						"key1": "key",
					},
				},
				{
					Source: "autotest",
					Name:   "custom",
					Labels: map[string]string{
						"key1": "key",
					},
				},
			},
		},
		{
			sourceSnippetConfigs: []apistructs.SnippetConfig{
				{
					Source: "local",
					Name:   "pipeline",
					Labels: map[string]string{
						"key3": "key",
					},
				},
			},
		},
		{
			sourceSnippetConfigs: []apistructs.SnippetConfig{
				{
					Source: "autotest",
					Name:   "custom",
					Labels: map[string]string{
						"key3": "key",
					},
				},
				{
					Source: "autotest",
					Name:   "custom",
					Labels: map[string]string{
						"key1": "key",
					},
				},
				{
					Source: "autotest",
					Name:   "custom",
					Labels: map[string]string{
						"key1": "key",
					},
				},
			},
		},
		{
			sourceSnippetConfigs: []apistructs.SnippetConfig{},
		},
		{
			sourceSnippetConfigs: nil,
		},
	}
	for _, data := range table {
		_, err := svc.handleQueryPipelineYamlBySnippetConfigs(data.sourceSnippetConfigs)
		assert.NoError(t, err)
	}
}
