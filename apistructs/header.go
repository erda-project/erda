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

// ErrorResponse 统一的 response 的 err 部分
type ErrorResponse struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Ctx  interface{} `json:"ctx"`
}

// Header 统一的 response 的除了接口数据的 header 部分
type Header struct {
	Success bool          `json:"success" `
	Error   ErrorResponse `json:"err"`
}

// RequestHeader 统一的 request
type RequestHeader struct {
	Locale string
}

// UserInfoHeader 由 openAPI 统一注入在 response 中
type UserInfoHeader struct {
	UserIDs  []string            `json:"userIDs"`
	UserInfo map[string]UserInfo `json:"userInfo"`
}

// EventHeader event 公共 header
type EventHeader struct {
	Event         string `json:"event"`
	Action        string `json:"action"`
	OrgID         string `json:"orgID"`
	ProjectID     string `json:"projectID"`
	ApplicationID string `json:"applicationID"`
	Env           string `json:"env"`
	// Content   PlaceHolder `json:"content"`
	TimeStamp string `json:"timestamp"`
}
