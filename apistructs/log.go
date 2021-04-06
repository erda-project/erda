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

type GetLogsResponse struct {
	Success bool      `json:"success"`
	Err     string    `json:"err"`
	Data    LogDetail `json:"data"`
}

type LogLine struct {
	Source    string `json:"source"`
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Offset    string `json:"offset"`
	Content   string `json:"content"`
	Level     string `json:"level"`
}

type LogDetail struct {
	Lines []LogLine `json:"lines"`
}

// LogPushRequest 推日志请求
type LogPushRequest struct {
	Lines []LogPushLine
}

// LogPushLine 推日志请求行
type LogPushLine struct {
	ID        string      `json:"id"`
	Source    string      `json:"source"`
	Timestamp int64       `json:"timestamp"`
	Content   string      `json:"content"`
	Stream    *string     `json:"stream,omitempty"`
	Offset    *int        `json:"offset,omitempty"`
	Tags      interface{} `json:"tags,omitempty"`
}

var (
	CollectorLogPushStreamStdout = "stdout"
	CollectorLogPushStreamStderr = "stderr"
)
