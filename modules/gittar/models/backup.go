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

package models

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar/pkg/gitmodule"
)

// 获取备份列表
func (svc *Service) GetBackupList(pageNo, pageSize int, repo *gitmodule.Repository) (*apistructs.BackupListResponse, error) {
	var res apistructs.BackupListResponse
	offset := (pageNo - 1) * pageSize
	err := svc.db.Table("dice_files").Select("*").
		Joins("LEFT OUTER JOIN dice_repo_files on dice_repo_files.uuid = dice_files.uuid").
		Joins("LEFT OUTER JOIN uc_user on uc_user.id = dice_files.creator").
		Where("dice_repo_files.repo_id = ?", repo.ID).
		Order("dice_files.created_at desc").
		Offset(offset).Limit(pageSize).Find(&res.RepoFiles).
		// reset offset & limit before count
		Offset(0).Limit(-1).Count(&res.Total).Error
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// 添加备份记录
func (svc *Service) AddBackupRecording(repo *gitmodule.Repository, commitID, remark string) (*apistructs.RepoFiles, error) {
	var files apistructs.File
	err := svc.db.Table("dice_files").Last(&files).Error
	if err != nil {
		return nil, err
	}
	res := apistructs.RepoFiles{
		RepoID:   repo.ID,
		Remark:   remark,
		UUID:     files.UUID,
		CommitID: commitID,
	}
	fmt.Println(res)
	err = svc.db.Table("dice_repo_files").Create(&res).Error
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// 删除备份记录
func (svc *Service) DeleteBackupRecording(uuid string) (*apistructs.RepoFiles, error) {
	var file apistructs.RepoFiles
	err := svc.db.Table("dice_repo_files").Where("uuid = ?", uuid).Delete(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}
