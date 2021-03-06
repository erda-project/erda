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

// Package apistructs 自动化运行脚本相关
package apistructs

// ScriptInfo 脚本信息
type ScriptInfo struct {
	Md5  string `json:"md5"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	// 脚本名逗号分隔，ALL代表终止全部脚本
	ScriptBlackList []string `json:"blackList"`
}

// 获取自动化运维脚本信息

// GetScriptInfoResponse GET /api/script/info
type GetScriptInfoResponse struct {
	Header
	Data ScriptInfo `json:"data"`
}

// TaskInfo 自动化运行任务信息
type TaskInfo struct {
	ScriptName  string `json:"scriptName"`
	ClusterName string `json:"clusterName"`
	Running     bool   `json:"running"`
	LastStatus  bool   `json:"lastStatus"`
	Md5         string `json:"md5"`
	StartAt     int64  `json:"startAt"`
	EndAt       int64  `json:"endAt"`
	LastError   string `json:"lastError"`
	ErrorAt     int64  `json:"errorAt"`
}

// GetTasksInfoResponse 前端获取任务运行状态列表返回
type GetTasksInfoResponse struct {
	Header
	Data []TaskInfo `json:"data"`
}

// RunScriptResponse 脚本执行结果
type RunScriptResponse struct {
	Header
	Data interface{} `json:"data"`
}
