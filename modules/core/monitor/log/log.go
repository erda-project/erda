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

package log

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
