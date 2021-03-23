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
