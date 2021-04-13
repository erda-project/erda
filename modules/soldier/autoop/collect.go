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

package autoop

type LogContent struct {
	Source    string `json:"source"`
	ID        string `json:"id"`
	Content   string `json:"content"`
	TimeStamp int64  `json:"timestamp"`
	Stream    string `json:"stream"`
	Tags      Tag    `json:"tags"`
}

type Tag struct {
	Level string `json:"level"`
}

func NewLogContent() LogContent {
	return LogContent{Source: "job"}
}
