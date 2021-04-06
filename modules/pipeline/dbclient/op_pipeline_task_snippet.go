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

package dbclient

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// UpdatePipelineTaskSnippetDetail 更新 snippet task 的 snippet 信息
func (client *Client) UpdatePipelineTaskSnippetDetail(id uint64, snippetDetail apistructs.PipelineTaskSnippetDetail, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).Cols("snippet_pipeline_detail").Update(&spec.PipelineTask{SnippetPipelineDetail: &snippetDetail})
	return err
}
