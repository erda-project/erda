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

package moreOperation

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type MoreOperation struct {
	CtxBdl protocol.ContextBundle

	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Props      map[string]interface{} `json:"props"`
	Operations map[string]interface{} `json:"operations"`
}

func (m *MoreOperation) marshal(c *apistructs.Component) error {

	propValue, err := json.Marshal(m.Props)
	if err != nil {
		return err
	}
	var props interface{}
	err = json.Unmarshal(propValue, &props)
	if err != nil {
		return err
	}

	c.Operations = m.Operations
	c.Props = props
	c.Type = m.Type
	c.Name = m.Name
	return nil
}

func (l *MoreOperation) Import(c *apistructs.Component) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, l); err != nil {
		return err
	}
	return nil
}

func (m *MoreOperation) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	if err := m.Import(c); err != nil {
		logrus.Errorf("import component failed, err:%v", err)
		return err
	}

	defer func() {
		fail := m.marshal(c)
		if err == nil && fail != nil {
			err = fail
		}
	}()
	m.Name = "moreOperation"
	m.Type = "Dropdown"
	m.Props = map[string]interface{}{
		"menus": []interface{}{
			map[string]interface{}{
				"key":   "import",
				"label": "导入",
			},
			map[string]interface{}{
				"key":   "record",
				"label": "导入导出记录",
			},
		},
	}
	m.Operations = map[string]interface{}{
		"import": map[string]interface{}{
			"key":    "import",
			"reload": false,
		},
		"record": map[string]interface{}{
			"key":    "record",
			"reload": false,
		},
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &MoreOperation{}
}
