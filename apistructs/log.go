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
