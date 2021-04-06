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

/*
对外提供 API 用于注册特定 label 对 其他 label 的关系
暂定这个特定 label 的 key 为 "REGISTERED_LABEL".(constant.RegisterLabelKey)

e.g.
PUT /dice/eventbox/register/<VALUE>
body: {"labels": map[string]string{"<label1>":"<value1>", "<label2>":"<value2>"}}

之后在发送消息的时候，带上 label : {"REGISTERED_LABEL":"<VALUE>"},
相当于 带上了 上面所注册的所有 labels

*/
package register
