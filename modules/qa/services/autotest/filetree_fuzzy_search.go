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

func (svc *Service) FuzzySearchFileTreeNodes(req apistructs.UnifiedFileTreeNodeFuzzySearchRequest) ([]apistructs.UnifiedFileTreeNode, error) {
	// 参数校验
	if err := req.BasicValidate(); err != nil {
		return nil, apierrors.ErrFuzzySearchAutoTestFileTreeNodes.InvalidParameter(err)
	}

	// 若起始父节点未指定，默认为 0
	if req.FromPinode == "" {
		req.FromPinode = rootDirNodePinode
	}

	// 计算 pinodes
	var pinodes []string
	// 若需要递归，则递归查询从起始节点开始的子目录 id 列表
	if req.Recursive {
		// 起始节点为根目录且递归，则无需递归查询子目录 id，直接用 scope & scopeID 过滤即可
		if req.FromPinode == rootDirNodePinode {
			pinodes = nil
		} else {
			// 加入指定的 pinode
			pinodes = append(pinodes, req.FromPinode)
			// 递归
			if err := svc.recursiveGetDirInodes(req.FromPinode, &pinodes); err != nil {
				return nil, apierrors.ErrFuzzySearchAutoTestFileTreeNodes.InternalError(err)
			}
		}
	}

	// 搜索
	dbNodes, err := svc.db.FuzzySearchAutoTestFileTreeNodes(req.Scope, req.ScopeID, req.PrefixFuzzy, req.SuffixFuzzy, req.Fuzzy, pinodes, req.CreatorID)
	if err != nil {
		return nil, apierrors.ErrFuzzySearchAutoTestFileTreeNodes.InternalError(err)
	}

	// 转换
	nodes := batchConvertToUnifiedFileTreeNodes(dbNodes...)

	return nodes, nil
}

func (svc *Service) recursiveGetDirInodes(pinode string, inodes *[]string) error {
	nodes, err := svc.db.ListAutoTestFileTreeNodeByPinode(pinode)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		// 忽略非目录
		if !node.Type.IsDir() {
			continue
		}
		*inodes = append(*inodes, node.Inode)

		if err := svc.recursiveGetDirInodes(node.Inode, inodes); err != nil {
			return err
		}
	}
	return nil
}
