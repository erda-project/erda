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

func (svc *Service) GetFileTreeNode(req apistructs.UnifiedFileTreeNodeGetRequest) (*apistructs.UnifiedFileTreeNode, error) {
	// 参数校验
	if err := req.BasicValidate(); err != nil {
		return nil, apierrors.ErrGetAutoTestFileTreeNode.InvalidParameter(err)
	}
	// 查询
	node, exist, err := svc.db.GetAutoTestFileTreeNodeByInode(req.Inode)
	if err != nil {
		return nil, apierrors.ErrGetAutoTestFileTreeNode.InternalError(err)
	}
	if !exist {
		return nil, apierrors.ErrGetAutoTestFileTreeNode.NotFound()
	}

	// 查询节点 meta
	meta, _, err := svc.db.GetAutoTestFileTreeNodeMetaByInode(node.Inode)
	if err != nil {
		return nil, apierrors.ErrGetAutoTestFileTreeNode.InternalError(err)
	}
	return convertToUnifiedFileTreeNode(node, meta), nil
}
