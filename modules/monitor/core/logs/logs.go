// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package logs

// Log .
type Log struct {
	Source    string            `json:"source"`
	ID        string            `json:"id"`
	Stream    string            `json:"stream"`
	Content   string            `json:"content"`
	Offset    int64             `json:"offset"`
	Timestamp int64             `json:"timestamp"`
	Tags      map[string]string `json:"tags"`
}

// LogMeta .
type LogMeta struct {
	Source string            `json:"source"`
	ID     string            `json:"id"`
	Tags   map[string]string `json:"tags"`
}
