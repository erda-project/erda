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

package nacos

import "github.com/erda-project/erda-proto-go/msp/configcenter/pb"

// SearchMode .
type SearchMode string

var (
	// SearchModeBlur .
	SearchModeBlur SearchMode = "BLUR"
	// SearchModeAccurate .
	SearchModeAccurate SearchMode = "ACCURATE"
)

// Adapter .
type Adapter struct {
	ClusterName string
	Addr        string
	User        string
	Password    string
}

// NewAdapter .
func NewAdapter(clusterName, addr, user, password string) *Adapter {
	return &Adapter{
		ClusterName: clusterName,
		Addr:        addr,
		User:        user,
		Password:    password,
	}
}

// SearchResponse .
type SearchResponse struct {
	Total       int64
	Pages       int64
	ConfigItems []*ConfigItem
}

// ConfigItem .
type ConfigItem struct {
	DataID  string
	Group   string
	Content string
}

// ToConfigCenterGroups .
func (s *SearchResponse) ToConfigCenterGroups() *pb.ConfigCenterGroups {
	return &pb.ConfigCenterGroups{}
}

// SearchConfig .
func (a *Adapter) SearchConfig(mode SearchMode, tenantName, groupName, dataID string, page, pageSize int) (*SearchResponse, error) {
	// TODO .
	return &SearchResponse{}, nil
}

// SaveConfig .
func (a *Adapter) SaveConfig(tenantName, groupName, dataID, content string) error {
	// TODO .
	return nil
}
