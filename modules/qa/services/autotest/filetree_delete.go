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
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

func (svc *Service) DeleteFileTreeNode(req apistructs.UnifiedFileTreeNodeDeleteRequest) (*apistructs.UnifiedFileTreeNode, error) {
	// 参数校验
	if err := req.BasicValidate(); err != nil {
		return nil, apierrors.ErrGetAutoTestFileTreeNode.InvalidParameter(err)
	}
	// 查询
	node, err := svc.GetFileTreeNode(apistructs.UnifiedFileTreeNodeGetRequest{
		Inode:        req.Inode,
		IdentityInfo: req.IdentityInfo,
	})
	if err != nil {
		return nil, apierrors.ErrDeleteAutoTestFileTreeNode.InvalidParameter(err)
	}

	// 删除历史记录
	svc.DeleteFileTreeNodeHistory(req.Inode)

	// 删除节点
	if err := svc.db.DeleteAutoTestFileTreeNodeByInode(req.Inode); err != nil {
		return nil, apierrors.ErrDeleteAutoTestFileTreeNode.InternalError(err)
	}

	// 递归操作
	go func() {
		if node.Type.IsDir() {
			subNodes, err := svc.db.ListAutoTestFileTreeNodeByPinode(req.Inode)
			if err != nil {
				logrus.Errorf("failed to list autotest sub nodes when delete dir node, inode: %s, err: %v", req.Inode, err)
			}
			for _, subNode := range subNodes {
				inode := subNode.Inode
				if _, err := svc.DeleteFileTreeNode(apistructs.UnifiedFileTreeNodeDeleteRequest{
					Inode:        inode,
					IdentityInfo: req.IdentityInfo,
				}); err != nil {
					logrus.Errorf("failed to recursive delete file tree node, inode: %s, err: %v", inode, err)
				}

				// 删除历史记录
				svc.DeleteFileTreeNodeHistory(inode)
			}
		}
	}()

	return node, nil
}
