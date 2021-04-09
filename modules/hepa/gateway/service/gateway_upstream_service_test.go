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

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/erda-project/erda/modules/hepa/config"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
)

func Test_safeChange(t *testing.T) {
	type args struct {
		apis   []orm.GatewayUpstreamApi
		change func(orm.GatewayUpstreamApi)
	}
	config.ServerConf.RegisterSliceSize = 10
	config.ServerConf.RegisterInterval = 0
	var apis []orm.GatewayUpstreamApi
	for i := 0; i < 105; i++ {
		apis = append(apis, orm.GatewayUpstreamApi{})
	}
	var newapis []orm.GatewayUpstreamApi
	change := func(api orm.GatewayUpstreamApi) {
		api.RegisterId = "1"
		newapis = append(newapis, api)
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			"test",
			args{apis, change},
		},
	}
	errhappen := false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safeChange(tt.args.apis, tt.args.change, &errhappen)
		})
	}
	assert.Equal(t, len(apis), len(newapis))
	for _, api := range newapis {
		assert.Equal(t, "1", api.RegisterId)
	}
}
