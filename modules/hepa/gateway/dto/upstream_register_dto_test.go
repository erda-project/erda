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

package dto

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestUpstreamRegisterDto_Init(t *testing.T) {
	jsonstr := `{"registerId":123456}`
	type fields struct {
		OldRegisterId int `json:"registerId"`
	}
	obj := &fields{}
	_ = json.Unmarshal([]byte(jsonstr), obj)
	fmt.Printf("%+v", obj)
}
