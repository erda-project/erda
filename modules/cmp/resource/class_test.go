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

package resource

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
	_interface "github.com/erda-project/erda/modules/cmp/cmp_interface"
)

func TestResource_GetClusterPie(t *testing.T) {
	type fields struct {
		Ctx    context.Context
		Server _interface.Provider
		I18N   i18n.Translator
		Lang   i18n.LanguageCodes
	}
	res := &pb.GetClusterResourcesResponse{
		List: []*pb.ClusterResourceDetail{{ClusterName: "terminus"}},
	}
	pie := &PieData{}
	pie.Series = append(pie.Series, PieSerie{
		Name: "distribution by cluster",
		Type: "pie",
		Data: []SerieData{{
			Value: 0,
			Name:  "terminus",
		},
		},
	})
	type args struct {
		resourceType string
		resources    *pb.GetClusterResourcesResponse
	}
	tests := []struct {
		name           string
		args           args
		wantProjectPie *PieData
		wantErr        bool
	}{
		{
			name: "test",
			args: args{
				resourceType: CPU,
				resources:    res,
			},
			wantProjectPie: pie,
		},

		{
			name: "test",
			args: args{
				resourceType: Memory,
				resources:    res,
			},
			wantProjectPie: pie,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resource{I18N: nopTranslator{}}
			gotProjectPie, err := r.GetClusterPie(tt.args.resourceType, tt.args.resources)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClusterPie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotProjectPie, tt.wantProjectPie) {
				t.Errorf("GetClusterPie() gotProjectPie = %v, want %v", gotProjectPie, tt.wantProjectPie)
			}
		})
	}
}

func TestResource_GetPrincipalPie(t *testing.T) {
	type fields struct {
		Ctx    context.Context
		Server _interface.Provider
		I18N   i18n.Translator
		Lang   i18n.LanguageCodes
	}
	type args struct {
		resourceType string
		resp         *apistructs.GetQuotaOnClustersResponse
	}
	pie := &PieData{}
	pie.Series = append(pie.Series, PieSerie{
		Name: "distribution by principal",
		Type: "pie",
		Data: []SerieData{{
			Value: 0,
		},
		},
	})
	resp := &apistructs.GetQuotaOnClustersResponse{
		Owners: []*apistructs.OwnerQuotaOnClusters{{ID: 1, Projects: []*apistructs.ProjectQuotaOnClusters{{ID: 1}}}},
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantPrincipalPie *PieData
		wantErr          bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			args: args{
				resourceType: CPU,
				resp:         resp,
			},
			wantPrincipalPie: pie,
		},
		{
			name: "test2",
			args: args{
				resourceType: Memory,
				resp:         resp,
			},
			wantPrincipalPie: pie,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resource{
				I18N: nopTranslator{},
			}
			_, err := r.GetPrincipalPie(tt.args.resourceType, tt.args.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPrincipalPie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestResource_GetProjectPie(t *testing.T) {
	type fields struct {
		Ctx    context.Context
		Server _interface.Provider
		I18N   i18n.Translator
		Lang   i18n.LanguageCodes
	}
	type args struct {
		resourceType string
		resp         *apistructs.GetQuotaOnClustersResponse
	}
	pie := &PieData{}
	pie.Series = append(pie.Series, PieSerie{
		Name: "distribution by project",
		Type: "pie",
	})
	resp := &apistructs.GetQuotaOnClustersResponse{
		Owners: []*apistructs.OwnerQuotaOnClusters{{ID: 1, Projects: []*apistructs.ProjectQuotaOnClusters{{ID: 1}}}},
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantProjectPie *PieData
		wantErr        bool
	}{
		{
			name: "test",
			args: args{
				resourceType: CPU,
				resp:         resp,
			},
			wantProjectPie: pie,
		},
		{
			name: "test",
			args: args{
				resourceType: Memory,
				resp:         resp,
			},
			wantProjectPie: pie,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resource{
				I18N: nopTranslator{},
			}
			_, err := r.GetProjectPie(tt.args.resourceType, tt.args.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjectPie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
