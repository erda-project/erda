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

package component_protocol

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

type X struct {
	State XState      `json:"state"`
	Props interface{} `json:"props"`
	Name  string      `json:"-"`
	Type  string
}

type XState struct {
	S string `json:"s"`
}

func XC(x *X) CompRender {
	return x
}

func (a *X) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, _ *apistructs.GlobalStateData) error {
	return nil
}

func Test_marshal(t *testing.T) {
	init := apistructs.Component{
		Version:    "1.0",
		Type:       "well",
		Name:       "good",
		Props:      map[string]interface{}{"old": "val"},
		Data:       map[string]interface{}{"k": "v"},
		State:      map[string]interface{}{"s": "l"},
		Operations: map[string]interface{}{"v": "n"},
	}
	x := X{}
	x1 := XC(&x)
	assert.Nil(t, unmarshal(&x1, &init))
	assert.Equal(t, "l", x.State.S)
	assert.Equal(t, map[string]interface{}{"old": "val"}, x.Props)
	assert.Equal(t, "", x.Name)
	assert.Equal(t, "well", x.Type)

	x.State.S = "changed"
	x.Props = map[string]interface{}{"new": "key"}
	x.Type = "sell"

	assert.Nil(t, marshal(&x1, &init))
	assert.Equal(t, "changed", init.State["s"])
	assert.Equal(t, map[string]interface{}{"new": "key"}, init.Props)

	assert.Equal(t, "sell", init.Type)
	// no changed
	assert.Equal(t, "good", init.Name)
}
