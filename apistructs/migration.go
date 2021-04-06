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

package apistructs

// MigrationStatus addon规格信息返回res
type MigrationStatusDesc struct {
	// Status 返回的运行状态
	Status StatusCode `json:"status"`
	// Desc 说明信息
	Desc string `json:"desc"`
}

const (
	// MigrationStatusInit migration初始化
	MigrationStatusInit string = "INIT"
	// MigrationStatusPending migration等待
	MigrationStatusPending string = "PENDING"
	// MigrationStatusRunning migration running
	MigrationStatusRunning string = "RUNNING"
	// MigrationStatusFail migration 失败
	MigrationStatusFail string = "FAIL"
	// MigrationStatusFinish migration 完成
	MigrationStatusFinish string = "FINISH"
	// MigrationStatusDeleted migration 删除
	MigrationStatusDeleted string = "DELETE"
)
