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

package subscriber

import (
	"github.com/erda-project/erda/modules/eventbox/types"
)

type Subscriber interface {
	// 各个实现自己解析 dest
	// 返回 []error , 是因为发送消息的目的可能是多个
	// dest: marshaled string
	// content: marshaled string
	Publish(dest string, content string, time int64, m *types.Message) []error
	Status() interface{}
	Name() string
}
