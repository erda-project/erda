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
	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

func (svc *Service) FindFileTreeNodeAncestors(req apistructs.UnifiedFileTreeNodeFindAncestorsRequest) ([]apistructs.UnifiedFileTreeNode, error) {
	// 参数校验
	if err := req.BasicValidate(); err != nil {
		return nil, apierrors.ErrFindAutoTestFileTreeNodeAncestors.InvalidParameter(err)
	}

	// 查询当前节点
	currentNode, err := svc.GetFileTreeNode(apistructs.UnifiedFileTreeNodeGetRequest{
		Inode:        req.Inode,
		IdentityInfo: req.IdentityInfo,
	})
	if err != nil {
		return nil, apierrors.ErrFindAutoTestFileTreeNodeAncestors.InvalidParameter(err)
	}

	// 递归查询父节点
	var dbAncestors []dao.AutoTestFileTreeNode
	if err := svc.recursivelyFindAncestors(currentNode.Pinode, &dbAncestors); err != nil {
		return nil, apierrors.ErrFindAutoTestFileTreeNodeAncestors.InternalError(err)
	}

	// 转换
	ancestors := batchConvertToUnifiedFileTreeNodes(dbAncestors...)
	ancestors = append(ancestors, *currentNode)

	return ancestors, nil
}

// 层层递归，按照目录顺序，根目录在前
// 递归结果 ancestors 需要调用方初始化传入
func (svc *Service) recursivelyFindAncestors(pinode string, ancestors *[]dao.AutoTestFileTreeNode) error {
	if pinode == rootDirNodePinode {
		return nil
	}
	pNode, exist, err := svc.db.GetAutoTestFileTreeNodeByInode(pinode)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("parent node not exist")
	}
	*ancestors = append([]dao.AutoTestFileTreeNode{*pNode}, *ancestors...)
	if err := svc.recursivelyFindAncestors(pNode.Pinode, ancestors); err != nil {
		return err
	}
	return nil
}
