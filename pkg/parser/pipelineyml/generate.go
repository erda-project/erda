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

package pipelineyml

import (
	"bytes"

	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/pkg/strutil"
)

// GenerateYml 根据 spec 重新生成 yaml 文本，一般用于对 spec 进行调整后重新生成 yaml 文本
func GenerateYml(s *Spec) ([]byte, error) {
	polishNamespaces(s)
	var newYmlBuf bytes.Buffer
	encoder := yaml.NewEncoder(&newYmlBuf)
	encoder.SetIndent(1)
	if err := encoder.Encode(s); err != nil {
		return nil, err
	}
	return newYmlBuf.Bytes(), nil
}

// polishNamespaces 遍历 actions，判断 namespaces 长度：
// 因为目前 namespaces 的策略是 alias 一定会注入，所以可以删除 alias，优化最终生成的 yaml。
func polishNamespaces(s *Spec) {
	for _, stage := range s.Stages {
		for _, typedAction := range stage.Actions {
			for _, action := range typedAction {
				if action == nil {
					action = &Action{}
				}
				action.Namespaces = strutil.RemoveSlice(action.Namespaces, action.Alias.String())
			}
		}
	}
}
