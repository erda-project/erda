// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package autotest

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apps/dop/services/apierrors"
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
