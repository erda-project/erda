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

package xmind

type Content []Sheet

type Sheet struct {
	Topic Topic `json:"topic,omitempty" xml:"topic,omitempty"`
}

type Topic struct {
	Title  string  `json:"title,omitempty" xml:"title,omitempty"`
	Topics []Topic `json:"topics,omitempty" xml:"topics,omitempty"`
}

func (t Topic) GetFirstSubTopicTitle() string {
	if len(t.Topics) == 0 {
		return ""
	}
	return t.Topics[0].Title
}
