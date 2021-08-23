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

package db

import "time"

const (
	TableSpTraceRequestHistory = "sp_trace_request_history"
)

type TraceRequestHistory struct {
	RequestId      string    `gorm:"column:request_id" db:"request_id" json:"request_id" form:"request_id"`
	Name           string    `gorm:"column:name" db:"name" json:"name" form:"name"`
	TerminusKey    string    `gorm:"column:terminus_key" db:"terminus_key" json:"terminus_key" form:"terminus_key"`
	Url            string    `gorm:"column:url" db:"url" json:"url" form:"url"`
	QueryString    string    `gorm:"column:query_string" db:"query_string" json:"query_string" form:"query_string"`
	Header         string    `gorm:"column:header" db:"header" json:"header" form:"header"`
	Body           string    `gorm:"column:body" db:"body" json:"body" form:"body"`
	Method         string    `gorm:"column:method" db:"method" json:"method" form:"method"`
	Status         int       `gorm:"column:status" db:"status" json:"status" form:"status"`
	ResponseStatus int       `gorm:"column:response_status" db:"response_status" json:"response_status" form:"response_status"`
	ResponseBody   string    `gorm:"column:response_body" db:"response_body" json:"response_body" form:"response_body"`
	CreateTime     time.Time `gorm:"column:create_time" db:"create_time" json:"create_time" form:"create_time"`
	UpdateTime     time.Time `gorm:"column:update_time" db:"update_time" json:"update_time" form:"update_time"`
}

func (TraceRequestHistory) TableName() string { return TableSpTraceRequestHistory }
