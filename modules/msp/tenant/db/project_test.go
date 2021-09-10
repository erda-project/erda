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

package db

import (
	"reflect"
	"testing"

	"github.com/jinzhu/gorm"
)

func TestMSPProjectDB_Query(t *testing.T) {
	mysqldb, _, _ := MockInit(MYSQL)
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *MSPProject
		wantErr bool
	}{
		{name: "case1", fields: fields{DB: mysqldb}, args: args{id: "1"}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &MSPProjectDB{
				DB: tt.fields.DB,
			}
			got, err := db.Query(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Query() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMSPProjectDB_Delete(t *testing.T) {
	mysqldb, _, _ := MockInit(MYSQL)
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *MSPProject
		wantErr bool
	}{
		{name: "case1", fields: fields{DB: mysqldb}, args: args{id: "1"}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &MSPProjectDB{
				DB: tt.fields.DB,
			}
			got, err := db.Delete(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Delete() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMSPProjectDB_Update(t *testing.T) {
	mysqldb, _, _ := MockInit(MYSQL)
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		project *MSPProject
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *MSPProject
		wantErr bool
	}{
		{name: "case1", fields: fields{DB: mysqldb}, args: args{project: nil}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &MSPProjectDB{
				DB: tt.fields.DB,
			}
			got, err := db.Update(tt.args.project)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Update() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMSPProjectDB_Create(t *testing.T) {
	mysqldb, _, _ := MockInit(MYSQL)
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		project *MSPProject
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *MSPProject
		wantErr bool
	}{
		{name: "case1", fields: fields{DB: mysqldb}, args: args{project: nil}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &MSPProjectDB{
				DB: tt.fields.DB,
			}
			got, err := db.Create(tt.args.project)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() got = %v, want %v", got, tt.want)
			}
		})
	}
}
