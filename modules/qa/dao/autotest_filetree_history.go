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

package dao

import "github.com/erda-project/erda/pkg/dbengine"

type AutoTestFileTreeNodeHistory struct {
	dbengine.BaseModel
	Pinode        string                        `gorm:"type:bigint(20)"` // root dir 的 pinode 为 "0"，表示无 pinode
	Inode         string                        `gorm:"type:bigint(20)"`
	Name          string                        `gorm:"name"`
	Desc          string                        `gorm:"desc"`
	CreatorID     string                        `gorm:"creator_id"`
	UpdaterID     string                        `gorm:"updater_id"`
	PipelineYml   string                        `gorm:"pipeline_yml"`   // 节点的 pipeline yml 文件，pipeline 可以通过 snippetAction 配置
	SnippetAction snippetActionType             `gorm:"snippet_action"` // 其他用例或计划通过 snippet 方式引用时当前节点时, 根据该参数拼装出 snippet action
	Extra         autoTestFileTreeNodeMetaExtra `gorm:"extra"`
}

func (AutoTestFileTreeNodeHistory) TableName() string {
	return "dice_autotest_filetree_nodes_histories"
}

func (db *DBClient) CreateAutoTestFileTreeNodeHistory(node *AutoTestFileTreeNodeHistory) error {
	return db.Create(node).Error
}

func (db *DBClient) DeleteAutoTestFileTreeNodeHistory(node *AutoTestFileTreeNodeHistory) error {
	var set AutoTestFileTreeNodeHistory
	return db.Where("id = ?", node.ID).Delete(set).Error
}

func (db *DBClient) ListAutoTestFileTreeNodeHistoryByinode(inode string) ([]AutoTestFileTreeNodeHistory, error) {
	var nodes []AutoTestFileTreeNodeHistory
	if err := db.Where("`inode` = ?", inode).Order("created_at desc").Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}
