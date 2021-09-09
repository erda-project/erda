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

package query

import (
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/modules/extensions/loghub/index/query/db"
)

func TestGetAllESClients_WithErrorAccessDb_Should_Return_Nil(t *testing.T) {
	p := &provider{
		db: &db.DB{
			LogDeployment: db.LogDeploymentDB{},
		},
	}

	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&p.db.LogDeployment), "List")
	monkey.PatchInstanceMethod(reflect.TypeOf(&p.db.LogDeployment), "List", func(_ *db.LogDeploymentDB) ([]*db.LogDeployment, error) {
		return nil, fmt.Errorf("boooooo!")
	})

	clients := p.getAllESClients()
	if clients != nil {
		t.Errorf("should return nil when fail to access logDeployment")
	}
}

func TestGetAllESClients_On_ExistsLogDeployment_Should_Return_None_Empty_Clients(t *testing.T) {
	p := &provider{
		db: &db.DB{
			LogDeployment: db.LogDeploymentDB{},
		},
	}

	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&p.db.LogDeployment), "List")
	monkey.PatchInstanceMethod(reflect.TypeOf(&p.db.LogDeployment), "List", func(_ *db.LogDeploymentDB) ([]*db.LogDeployment, error) {
		return []*db.LogDeployment{
			{ID: 123},
		}, nil
	})

	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(p), "getESClientsFromLogAnalyticsByLogDeployment")
	monkey.PatchInstanceMethod(reflect.TypeOf(p), "getESClientsFromLogAnalyticsByLogDeployment", func(_ *provider, addon string, logDeployments ...*db.LogDeployment) []*ESClient {
		return []*ESClient{
			&ESClient{URLs: "success"},
		}
	})

	clients := p.getAllESClients()
	if len(clients) == 0 {
		t.Errorf("should return non-empty ESClients list when exists logDeployment")
	}
}
