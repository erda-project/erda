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

package autotest

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

const (
	metaKeyPipelineYml = "pipelineYml"
)

// QueryPipelineSnippetYaml 提供协议接口供 pipeline 引擎在运行时调用，根据 snippet 配置查询 snippet 对应的 pipelineYaml 文件内容
// 规则：
//   - source: autotest
//   - name: file tree node inode
//   - labels: no need now
func (svc *Service) QueryPipelineSnippetYaml(req apistructs.SnippetConfig, identityIndo apistructs.IdentityInfo) (string, error) {
	// 校验 source
	if req.Source != apistructs.PipelineSourceAutoTest.String() {
		return "", apierrors.ErrQueryPipelineSnippetYaml.InvalidParameter(fmt.Errorf("invalid source"))
	}
	// 查询 node
	node, err := svc.GetFileTreeNode(apistructs.UnifiedFileTreeNodeGetRequest{
		Inode:        req.Name,
		IdentityInfo: identityIndo,
	})
	if err != nil {
		return "", apierrors.ErrQueryPipelineSnippetYaml.InternalError(err)
	}
	y, ok := node.Meta[metaKeyPipelineYml]
	if !ok {
		return "", apierrors.ErrQueryPipelineSnippetYaml.InternalError(fmt.Errorf("node doesn't have pipelineYml"))
	}
	return fmt.Sprintf("%v", y), nil
}
