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
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/uuid"
)

func (svc *Service) CreateFileTreeNode(req apistructs.UnifiedFileTreeNodeCreateRequest) (*apistructs.UnifiedFileTreeNode, error) {
	// 创建前校验
	node, err := svc.validateNodeBeforeCreate(req)
	if err != nil {
		return nil, apierrors.ErrCreateAutoTestFileTreeNode.InvalidParameter(err)
	}
	// 创建
	if err := svc.db.CreateAutoTestFileTreeNode(node); err != nil {
		return nil, apierrors.ErrCreateAutoTestFileTreeNode.InternalError(err)
	}
	// 转换
	return convertToUnifiedFileTreeNode(node, nil), nil
}

func (svc *Service) validateNodeBeforeCreate(req apistructs.UnifiedFileTreeNodeCreateRequest) (*dao.AutoTestFileTreeNode, error) {
	// 构造 node
	node := dao.AutoTestFileTreeNode{
		Type:      req.Type,
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
		Pinode:    req.Pinode,
		Name:      req.Name,
		Desc:      req.Desc,
		CreatorID: req.IdentityInfo.UserID,
		UpdaterID: req.IdentityInfo.UserID,
	}
	// 参数校验
	if !req.Type.Valid() {
		return nil, fmt.Errorf("invalid node type: %s", req.Type.String())
	}
	if req.Type.IsDir() {
		if req.Pinode == "" {
			// root dir
			if err := req.ValidateRootDir(); err != nil {
				return nil, err
			}
			// 根节点 pinode 设置为 0，而不是类 Linux pinode=inode，方便数据库插入和查找
			node.Pinode = rootDirNodePinode
		} else {
			// non-root dir
			if err := req.ValidateNonRootDir(); err != nil {
				return nil, err
			}
		}
	}
	if req.Type.IsFile() {
		if err := req.ValidateFile(); err != nil {
			return nil, err
		}
	}
	// 字段最大长度校验
	if err := strutil.Validate(node.Name, strutil.MaxLenValidator(maxSetNameLen)); err != nil {
		return nil, err
	}
	if err := strutil.Validate(node.Desc, strutil.MaxLenValidator(maxSetDescLen)); err != nil {
		return nil, err
	}

	if node.Pinode == rootDirNodePinode {
		// 根目录，一个 scope id 下有且只有一个根目录。若已存在则报错
		rootSet, exist, err := svc.db.GetAutoTestFileTreeScopeRootDir(node.Scope, node.ScopeID)
		if err != nil {
			return nil, err
		}
		if exist {
			return nil, fmt.Errorf("root dir already exist, inode: %s", rootSet.Inode)
		}
	} else {
		// 有父节点，校验名称是否重名
		pNode, exist, err := svc.db.GetAutoTestFileTreeNodeByInode(node.Pinode)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, fmt.Errorf("parent node not exist")
		}
		// 文件节点的 pNode 类型需要是 dir
		if node.Type.IsFile() && !pNode.Type.IsDir() {
			return nil, fmt.Errorf("file node's parent node type must be dir")
		}
		// 查询 pinode 下同名 inode 是否存在，若存在，则加(n)
		ensuredName, err := svc.ensureNodeName(node.Pinode, node.Name)
		if err != nil {
			return nil, err
		}
		node.Name = ensuredName
		// 从 pinode 继承 scope & scopeID
		node.Scope = pNode.Scope
		node.ScopeID = pNode.ScopeID
	}

	// 分配 inode
	node.Inode = uuid.SnowFlakeID()
	return &node, nil
}
