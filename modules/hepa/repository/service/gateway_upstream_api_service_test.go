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
	"reflect"
	"testing"

	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/repository/orm"

	"github.com/stretchr/testify/assert"
	"github.com/xormplus/xorm"
)

func TestNewGatewayUpstreamApiServiceImpl(t *testing.T) {
	tests := []struct {
		name    string
		want    *GatewayUpstreamApiServiceImpl
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGatewayUpstreamApiServiceImpl()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGatewayUpstreamApiServiceImpl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGatewayUpstreamApiServiceImpl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGatewayUpstreamApiServiceImpl_Insert(t *testing.T) {
	type fields struct {
		engine *orm.OrmEngine
	}
	type args struct {
		session *xorm.Session
		item    *orm.GatewayUpstreamApi
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &GatewayUpstreamApiServiceImpl{
				engine: tt.fields.engine,
			}
			got, err := impl.Insert(tt.args.session, tt.args.item)
			if (err != nil) != tt.wantErr {
				t.Errorf("GatewayUpstreamApiServiceImpl.Insert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GatewayUpstreamApiServiceImpl.Insert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGatewayUpstreamApiServiceImpl_updateFields(t *testing.T) {
	type fields struct {
		engine *orm.OrmEngine
	}
	type args struct {
		update *orm.GatewayUpstreamApi
		fields []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &GatewayUpstreamApiServiceImpl{
				engine: tt.fields.engine,
			}
			if err := impl.updateFields(tt.args.update, tt.args.fields...); (err != nil) != tt.wantErr {
				t.Errorf("GatewayUpstreamApiServiceImpl.updateFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGatewayUpstreamApiServiceImpl_GetLastApiId(t *testing.T) {
	engine, error := orm.GetSingleton()
	assert.Nil(t, error)
	type fields struct {
		engine *orm.OrmEngine
	}
	type args struct {
		cond *orm.GatewayUpstreamApi
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			"case1",
			fields{engine},
			args{&orm.GatewayUpstreamApi{
				UpstreamId: "0436d82d632644db9ba3917077d17e5c",
				ApiName:    "myapi",
			}},
			"b8ca6c5d61924c19befe826e3d0cc0a0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &GatewayUpstreamApiServiceImpl{
				engine: tt.fields.engine,
			}
			if got := impl.GetLastApiId(tt.args.cond); got != tt.want {
				t.Errorf("GatewayUpstreamApiServiceImpl.GetLastApiId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGatewayUpstreamApiServiceImpl_UpdateApiId(t *testing.T) {
	type fields struct {
		engine *orm.OrmEngine
	}
	type args struct {
		update *orm.GatewayUpstreamApi
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &GatewayUpstreamApiServiceImpl{
				engine: tt.fields.engine,
			}
			if err := impl.UpdateApiId(tt.args.update); (err != nil) != tt.wantErr {
				t.Errorf("GatewayUpstreamApiServiceImpl.UpdateApiId() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGatewayUpstreamApiServiceImpl_countInIds(t *testing.T) {
	type fields struct {
		engine *orm.OrmEngine
	}
	type args struct {
		ids []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &GatewayUpstreamApiServiceImpl{
				engine: tt.fields.engine,
			}
			got, err := impl.countInIds(tt.args.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("GatewayUpstreamApiServiceImpl.countInIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GatewayUpstreamApiServiceImpl.countInIds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGatewayUpstreamApiServiceImpl_GetPage(t *testing.T) {
	type fields struct {
		engine *orm.OrmEngine
	}
	type args struct {
		ids  []string
		page *common.Page
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *common.PageQuery
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &GatewayUpstreamApiServiceImpl{
				engine: tt.fields.engine,
			}
			got, err := impl.GetPage(tt.args.ids, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("GatewayUpstreamApiServiceImpl.GetPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GatewayUpstreamApiServiceImpl.GetPage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGatewayUpstreamApiServiceImpl_SelectInIds(t *testing.T) {
	type fields struct {
		engine *orm.OrmEngine
	}
	type args struct {
		ids []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []orm.GatewayUpstreamApi
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &GatewayUpstreamApiServiceImpl{
				engine: tt.fields.engine,
			}
			got, err := impl.SelectInIds(tt.args.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("GatewayUpstreamApiServiceImpl.SelectInIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GatewayUpstreamApiServiceImpl.SelectInIds() = %v, want %v", got, tt.want)
			}
		})
	}
}
