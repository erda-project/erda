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

	"github.com/xormplus/xorm"

	"github.com/erda-project/erda/modules/hepa/repository/orm"
)

func TestGatewayKongInfoServiceImpl_acquireKongAddr(t *testing.T) {
	type fields struct {
		engine        *orm.OrmEngine
		SessionHelper *SessionHelper
		executor      xorm.Interface
	}
	type args struct {
		netportalUrl string
		selfAz       string
		info         *orm.GatewayKongInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			"case1",
			fields{
				nil,
				nil,
				nil,
			},
			args{
				"inet://abc?ssl=on",
				"xxx",
				&orm.GatewayKongInfo{
					Az:       "xxx",
					KongAddr: "kong.local",
				},
			},
			"kong.local",
			false,
		},
		{
			"case2",
			fields{
				nil,
				nil,
				nil,
			},
			args{
				"inet://abc?ssl=on",
				"xxx",
				&orm.GatewayKongInfo{
					Az:       "yyy",
					KongAddr: "kong.local",
				},
			},
			"inet://abc?ssl=on/kong.local",
			false,
		},
		{
			"case3",
			fields{
				nil,
				nil,
				nil,
			},
			args{
				"inet://abc?ssl=on",
				"xxx",
				&orm.GatewayKongInfo{
					Az:       "yyy",
					KongAddr: "inet://cde/kong.local",
				},
			},
			"inet://abc?ssl=on/kong.local",
			false,
		},
		{
			"case4",
			fields{
				nil,
				nil,
				nil,
			},
			args{
				"inet://abc?ssl=on",
				"xxx",
				&orm.GatewayKongInfo{
					Az:       "xxx",
					KongAddr: "inet://ads/kong.local",
				},
			},
			"kong.local",
			false,
		},
		{
			"case5",
			fields{
				nil,
				nil,
				nil,
			},
			args{
				"inet://abc?ssl=on",
				"xxx",
				&orm.GatewayKongInfo{
					Az:       "yyy",
					KongAddr: "http://kong.local",
				},
			},
			"inet://abc?ssl=on/kong.local",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &GatewayKongInfoServiceImpl{
				engine:        tt.fields.engine,
				SessionHelper: tt.fields.SessionHelper,
				executor:      tt.fields.executor,
			}
			got, err := impl.acquireKongAddr(tt.args.netportalUrl, tt.args.selfAz, tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("GatewayKongInfoServiceImpl.acquireKongAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GatewayKongInfoServiceImpl.acquireKongAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}
