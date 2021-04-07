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

type EventBoxRequest struct {
	Sender  string                 `json:"sender"`
	Content interface{}            `json:"content"`
	Labels  map[string]interface{} `json:"labels"`
}

type EventBoxResponse struct {
	Header
}
type EventBoxGroupNotifyRequest struct {
	Sender        string
	GroupID       int64
	NotifyItem    *NotifyItem
	Channels      string
	NotifyContent *GroupNotifyContent
	Params        map[string]string
}
