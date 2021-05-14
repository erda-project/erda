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
	"github.com/erda-project/erda/modules/qa/dao"
)

func (svc *Service) CreateFileTreeNodeHistory(inode string) error {

	// 查询数量，超过 10 个的全部删除, 这里异步删除，查询是按照创建时间倒序
	histories, err := svc.db.ListAutoTestFileTreeNodeHistoryByinode(inode)
	if err != nil {
		return err
	}
	// 异步删除
	if len(histories) >= 10 {
		for i := 9; i < len(histories); i++ {
			go func(i int) {
				if err := svc.db.DeleteAutoTestFileTreeNodeHistory(&histories[i]); err != nil {
					logrus.Errorf("delete > 10 history %v error: %v", histories[i], err)
				}
			}(i)
		}
	}

	// 异步的去添加
	go func() {
		// 获取当前 inode 的详情
		var req = apistructs.UnifiedFileTreeNodeGetRequest{
			Inode: inode,
		}

		// 查询
		node, exist, err := svc.db.GetAutoTestFileTreeNodeByInode(req.Inode)
		if err != nil {
			logrus.Errorf("save fileTree node history error: query node %s info error: %v", inode, err)
			return
		}
		if !exist {
			logrus.Errorf("save fileTree node history error: query node %s info error: %v", inode, err)
			return
		}

		// 查询节点 meta
		meta, _, err := svc.db.GetAutoTestFileTreeNodeMetaByInode(node.Inode)
		if err != nil {
			logrus.Errorf("save fileTree node history error: query node %s meta error: %v", inode, err)
			return
		}

		// 转化为 history
		var history = dao.AutoTestFileTreeNodeHistory{
			Pinode:    node.Pinode,
			Inode:     node.Inode,
			CreatorID: node.CreatorID,
			UpdaterID: node.UpdaterID,
			Name:      node.Name,
			Desc:      node.Desc,
		}
		if meta != nil {
			history.PipelineYml = meta.PipelineYml
			history.SnippetAction = meta.SnippetAction
			history.Extra = meta.Extra
		}
		// 创建历史记录
		if err := svc.db.CreateAutoTestFileTreeNodeHistory(&history); err != nil {
			logrus.Errorf("save fileTree node %s history error: %v", inode, err)
		}
	}()
	return nil
}
