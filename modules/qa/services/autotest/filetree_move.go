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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

func (svc *Service) MoveFileTreeNode(req apistructs.UnifiedFileTreeNodeMoveRequest) (*apistructs.UnifiedFileTreeNode, error) {
	// 参数校验
	if err := req.BasicValidate(); err != nil {
		return nil, apierrors.ErrMoveAutoTestFileTreeNode.InvalidParameter(err)
	}
	// 查询 node
	node, err := svc.GetFileTreeNode(apistructs.UnifiedFileTreeNodeGetRequest{
		Inode:        req.Inode,
		IdentityInfo: req.IdentityInfo,
	})
	if err != nil {
		return nil, apierrors.ErrMoveAutoTestFileTreeNode.InvalidParameter(err)
	}
	// 目标目录无变化，无需移动
	if node.Pinode == req.Pinode {
		return node, nil
	}
	// 查询 pinode
	_, err = svc.GetFileTreeNode(apistructs.UnifiedFileTreeNodeGetRequest{
		Inode:        req.Pinode,
		IdentityInfo: req.IdentityInfo,
	})
	if err != nil {
		return nil, apierrors.ErrMoveAutoTestFileTreeNode.InvalidParameter(err)
	}
	// 目标 node 不能是当前 node 的子节点
	targetIsSubNode, err := svc.findNodeUnderTargetNode(req.Pinode, req.Inode)
	if err != nil {
		return nil, apierrors.ErrCopyAutoTestFileTreeNode.InternalError(err)
	}
	if targetIsSubNode {
		return nil, apierrors.ErrCopyAutoTestFileTreeNode.InvalidParameter(fmt.Errorf("cannot move to a sub node"))
	}
	// 移动 -> 更新 node.pinode，并校验 name
	ensuredName, err := svc.ensureNodeName(req.Pinode, node.Name)
	if err != nil {
		return nil, apierrors.ErrCopyAutoTestFileTreeNode.InternalError(err)
	}
	// 创建历史记录
	if err := svc.CreateFileTreeNodeHistory(req.Inode); err != nil {
		logrus.Errorf("node id %s history create error: %v", req.Inode, err)
	}
	// 移动pinode
	if err := svc.db.MoveAutoTestFileTreeNode(req.Inode, req.Pinode, ensuredName, req.IdentityInfo.UserID); err != nil {
		return nil, apierrors.ErrMoveAutoTestFileTreeNode.InvalidParameter(err)
	}
	// 查询
	movedNode, err := svc.GetFileTreeNode(apistructs.UnifiedFileTreeNodeGetRequest{
		Inode:        req.Inode,
		IdentityInfo: req.IdentityInfo,
	})
	if err != nil {
		return nil, apierrors.ErrMoveAutoTestFileTreeNode.InternalError(err)
	}
	return movedNode, nil
}
