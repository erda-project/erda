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

package adapt

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/cql"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/modules/pkg/bundle-ex/cmdb"
	"github.com/erda-project/erda/pkg/encoding/jsonmap"
)

func TestAdapt_newTicketAlertNotify(t *testing.T) {
	type fields struct {
		l logs.Logger
		//metricq                Queryer
		t    i18n.Translator
		db   *db.DB
		cql  *cql.Cql
		bdl  *bundle.Bundle
		cmdb *cmdb.Cmdb
		//dashboardAPI           DashboardAPI
		orgFilterTags          map[string]bool
		microServiceFilterTags map[string]bool
		silencePolicies        map[string]bool
	}
	type args struct {
		alertID uint64
		silence *pb.AlertNotifySilence
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *db.AlertNotify
	}{
		//{
		//	name: "test_newTicketAlertNotify",
		//	fields: fields{
		//		silencePolicies: map[string]bool{
		//			"silence": true,
		//		},
		//	},
		//	args: args{
		//		alertID: 11,
		//		silence: &AlertNotifySilence{
		//			Value:  5,
		//			Unit:   "second",
		//			Policy: "silence",
		//		},
		//	},
		//	want: nil,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapt{
				l: tt.fields.l,
				//metricq:                tt.fields.metricq,
				t:    tt.fields.t,
				db:   tt.fields.db,
				cql:  tt.fields.cql,
				bdl:  tt.fields.bdl,
				cmdb: tt.fields.cmdb,
				//dashboardAPI:           tt.fields.dashboardAPI,
				orgFilterTags:          tt.fields.orgFilterTags,
				microServiceFilterTags: tt.fields.microServiceFilterTags,
				silencePolicies:        tt.fields.silencePolicies,
			}
			if got := a.newTicketAlertNotify(tt.args.alertID, tt.args.silence); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTicketAlertNotify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdapt_compareNotify(t *testing.T) {
	type fields struct {
		l logs.Logger
		//metricq                Queryer
		t    i18n.Translator
		db   *db.DB
		cql  *cql.Cql
		bdl  *bundle.Bundle
		cmdb *cmdb.Cmdb
		//dashboardAPI           DashboardAPI
		orgFilterTags          map[string]bool
		microServiceFilterTags map[string]bool
		silencePolicies        map[string]bool
	}
	type args struct {
		a *db.AlertNotify
		b *db.AlertNotify
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "test_compareNotify",
			fields: fields{},
			args: args{
				a: &db.AlertNotify{
					NotifyTarget: jsonmap.JSONMap{},
				},
				b: &db.AlertNotify{
					NotifyTarget: jsonmap.JSONMap{},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := &Adapt{
				l: tt.fields.l,
				//metricq:                tt.fields.metricq,
				t:    tt.fields.t,
				db:   tt.fields.db,
				cql:  tt.fields.cql,
				bdl:  tt.fields.bdl,
				cmdb: tt.fields.cmdb,
				//dashboardAPI:           tt.fields.dashboardAPI,
				orgFilterTags:          tt.fields.orgFilterTags,
				microServiceFilterTags: tt.fields.microServiceFilterTags,
				silencePolicies:        tt.fields.silencePolicies,
			}
			if got := ad.compareNotify(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("compareNotify() = %v, want %v", got, tt.want)
			}
		})
	}
}
