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
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

func (svc *Service) ListFileTreeNodes(req apistructs.UnifiedFileTreeNodeListRequest) ([]apistructs.UnifiedFileTreeNode, error) {
	// 参数校验
	if err := req.BasicValidate(); err != nil {
		return nil, apierrors.ErrGetAutoTestFileTreeNode.InvalidParameter(err)
	}
	// 查询
	// root-dir
	if req.Pinode == rootDirNodePinode {
		rootDirNode, exist, err := svc.db.GetAutoTestFileTreeScopeRootDir(req.Scope, req.ScopeID)
		if err != nil {
			return nil, apierrors.ErrListAutoTestFileTreeNodes.InternalError(err)
		}
		if !exist {
			return nil, nil
		}
		return batchConvertToUnifiedFileTreeNodes(*rootDirNode), nil
	}
	// 普通 pinode
	nodes, err := svc.db.ListAutoTestFileTreeNodeByPinode(req.Pinode)
	if err != nil {
		return nil, apierrors.ErrListAutoTestFileTreeNodes.InternalError(err)
	}
	return batchConvertToUnifiedFileTreeNodes(nodes...), nil
}
