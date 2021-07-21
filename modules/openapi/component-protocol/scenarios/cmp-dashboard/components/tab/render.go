package tab

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

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

type operationData struct {
	Meta meta `json:"meta"`
}

type meta struct {
	Target RowData `json:"target"`
}

const (
	DefaultPageSize = 15
	DefaultPageNo   = 1
)

type Operation struct {
	Key      string            `json:"key"`
	Reload   bool              `json:"reload"`
	FillMeta string            `json:"fillMeta"`
	Meta     map[string]string `json:"meta"`
}

type dataOperation struct {
	Key         string                 `json:"key"`
	Reload      bool                   `json:"reload"`
	Text        string                 `json:"text"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip,omitempty"`
	Confirm     string                 `json:"confirm,omitempty"`
	Meta        interface{}            `json:"meta,omitempty"`
	Command     map[string]interface{} `json:"command,omitempty"`
}

type RowData struct {
	Name              string `json:"name"`
	SnippetPipelineID uint64 `json:"snippetPipelineID"`
}

type AutoTestRunStep struct {
	ApiSpec  map[string]interface{} `json:"apiSpec"`
	WaitTime int64                  `json:"waitTime"`
}

func (t *SteveTab) Import(c *apistructs.Component) error {
	var (
		b   []byte
		err error
	)
	if b, err = json.Marshal(c); err != nil {
		return err
	}
	if err = json.Unmarshal(b, t); err != nil {
		return err
	}
	return nil
}
func (t *SteveTab) getProps() {
	t.Props.TabMenu = make([]MenuPair, 2)
	t.Props.TabMenu[0] = MenuPair{
		key:  CPU_TAB,
		name: CPU_TAB_ZH,
	}
	t.Props.TabMenu[1] = MenuPair{
		key:  MEM_TAB,
		name: MEM_TAB_ZH,
	}
}
func GetOpsInfo(opsData interface{}) (map[string]string, error) {
	if opsData == nil {
		err := fmt.Errorf("empty operation data")
		return nil, err
	}
	var op Operation
	cont, err := json.Marshal(opsData)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", opsData, err)
		return nil, err
	}
	err = json.Unmarshal(cont, &op)
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return nil, err
	}
	return op.Meta, nil
}

func (t *SteveTab) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	// import components data
	if err := t.Import(c); err != nil {
		logrus.Errorf("import components failed, err:%v", err)
		return err
	}
	switch event.Operation {
	case apistructs.OnChangeOperation:
		if t.State.ActiveKey == CPU_TAB{

		}
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}
	return nil
}

func getOperations(currentTab string) map[string]interface{} {
	return map[string]interface{}{
		"onChange": Operation{
			Key:      "changeTab",
			Reload:   true,
			FillMeta: "activeKey",
			Meta: map[string]string{"activeKey":currentTab},
		},
	}
}

func RenderCreator() protocol.CompRender {
	st := &SteveTab{
		Type:       "Tabs",
		Operations: nil,
	}
	st.getProps()
	st.Operations = getOperations("cpu")
	return st
}
