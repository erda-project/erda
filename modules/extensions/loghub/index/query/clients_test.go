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

/*
func TestGetAllESClients_On_ExistsLogDeployment_Should_Return_None_Empty_Clients(t *testing.T) {
	p := provider{
		db: &db.DB{
			LogDeployment: db.LogDeploymentDB{},
		},
		timeRanges: make(map[string]map[string]*timeRange),
		reload:     make(chan struct{}),
	}

	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&p.db.LogDeployment), "List")
	monkey.PatchInstanceMethod(reflect.TypeOf(&p.db.LogDeployment), "List", func(_ *db.LogDeploymentDB) ([]*db.LogDeployment, error) {
		return []*db.LogDeployment{
			&db.LogDeployment{
				ClusterName:  "cluster_1",
				ClusterType:  0,
				ESURL:        "http://localhost:9200",
				ESConfig:     "{}",
				CollectorURL: "http://collector:7096",
			},
		}, nil
	})

	// why can not patch the provider struct?
	//defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&p), "getESClientsFromLogAnalyticsByLogDeployment")
	//monkey.PatchInstanceMethod(reflect.TypeOf(&p), "getESClientsFromLogAnalyticsByLogDeployment", func(_ *provider, addon string, logDeployments []*db.LogDeployment) []*ESClient {
	//	return []*ESClient{
	//		&ESClient{URLs: "success"},
	//	}
	//})

	clients := p.getAllESClients()
	if len(clients) == 0 {
		t.Errorf("should return non-empty ESClients list when exists logDeployment")
	}
}

func TestGetESClientsFromLogAnalyticsByLogDeployment_On_Preload_Enabled_Should_Try_Fill_ESClient_Entrys(t *testing.T) {
	p := &provider{
		db: &db.DB{
			LogDeployment: db.LogDeploymentDB{},
		},
		C: &config{
			IndexPreload: true,
		},
		timeRanges: make(map[string]map[string]*timeRange),
		reload:     make(chan struct{}),
	}
	p.indices.Store(map[string]map[string][]*IndexEntry{
		"cluster_1": map[string][]*IndexEntry{
			"addon_1": []*IndexEntry{
				&IndexEntry{Index: "rlogs-addon_1-2020.34-000001",
					Name: "addon_1",
				},
			},
		},
	})

	logDeployments := []*db.LogDeployment{
		&db.LogDeployment{
			ClusterName:  "cluster_1",
			ClusterType:  0,
			ESURL:        "http://localhost:9200",
			ESConfig:     "{}",
			CollectorURL: "http://collector:7096",
		},
	}

	clients := p.getESClientsFromLogAnalyticsByLogDeployment("addon_1", logDeployments...)
	if len(clients) == 0 || len(clients[0].Entrys) == 0 {
		t.Errorf("ESClient.Entrys should not empty when preload matched")
	}
}
*/
