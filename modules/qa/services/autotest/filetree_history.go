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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

func (svc *Service) QueryFileTreeNodeHistory(req apistructs.UnifiedFileTreeNodeHistorySearchRequest) ([]*apistructs.UnifiedFileTreeNode, error) {

	results, err := svc.db.ListAutoTestFileTreeNodeHistoryByinode(req.Inode)
	if err != nil {
		return nil, err
	}

	// 查询
	node, exist, err := svc.db.GetAutoTestFileTreeNodeByInode(req.Inode)
	if err != nil {
		return nil, apierrors.ErrGetAutoTestFileTreeNode.InternalError(err)
	}
	if !exist {
		return nil, apierrors.ErrGetAutoTestFileTreeNode.NotFound()
	}

	var historyInfos []*apistructs.UnifiedFileTreeNode
	for _, v := range results {
		historyInfos = append(historyInfos, historyConvertToUnifiedFileTreeNode(&v, node))
	}

	return historyInfos, nil
}

func historyConvertToUnifiedFileTreeNode(history *dao.AutoTestFileTreeNodeHistory, node *dao.AutoTestFileTreeNode) *apistructs.UnifiedFileTreeNode {
	nodeInfo := &apistructs.UnifiedFileTreeNode{
		Inode:     history.Inode,
		Pinode:    history.Pinode,
		Name:      history.Name,
		Desc:      history.Desc,
		CreatorID: history.CreatorID,
		UpdaterID: history.UpdaterID,
		CreatedAt: history.CreatedAt,
		UpdatedAt: history.UpdatedAt,
		Meta: map[string]interface{}{
			apistructs.AutoTestFileTreeNodeMetaKeyHistoryID:     history.ID,
			apistructs.AutoTestFileTreeNodeMetaKeyPipelineYml:   history.PipelineYml,
			apistructs.AutoTestFileTreeNodeMetaKeySnippetAction: history.SnippetAction,
			apistructs.AutoTestFileTreeNodeMetaKeyExtra:         history.Extra,
		},
	}

	if node != nil {
		nodeInfo.Scope = node.Scope
		nodeInfo.ScopeID = node.ScopeID
		nodeInfo.Type = node.Type
	}

	return nodeInfo
}
