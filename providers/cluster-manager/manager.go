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

package manager

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/clientgo"
	"github.com/erda-project/erda/pkg/clientgo/restclient"
	"github.com/erda-project/erda/pkg/httputil"
)

// GetClusterClientSet Get cluster client-go clientSet
func (p *provider) GetClusterClientSet(clusterName string) (*clientgo.ClientSet, error) {
	ci, err := p.bundle.GetCluster(clusterName)
	if err != nil {
		return nil, err
	}

	if ci.ManageConfig == nil {
		if ci.SchedConfig == nil {
			return nil, fmt.Errorf("manage config and scheduler config is nil")
		}
		return clientgo.New(ci.SchedConfig.MasterURL)
	}

	switch ci.ManageConfig.Type {
	case apistructs.ManageProxy:
		rc, err := restclient.GetDialerRestConfig(clusterName, ci.ManageConfig)
		if err != nil {
			return nil, err
		}
		return clientgo.NewClientSet(rc)
	case apistructs.ManageToken, apistructs.ManageCert:
		rc, err := restclient.GetRestConfig(ci.ManageConfig)
		if err != nil {
			return nil, err
		}
		return clientgo.NewClientSet(rc)
	case apistructs.ManageInet:
		fallthrough
	default:
		if ci.SchedConfig == nil {
			return nil, fmt.Errorf("scheduler config is nil when use inet/default type")
		}
		return clientgo.New(ci.SchedConfig.MasterURL)
	}
}

// GetCluster
func (p *provider) GetCluster(clusterName string) (*apistructs.ClusterInfo, error) {
	return p.bundle.GetCluster(clusterName)
}

// CreateCluster Create cluster with bundle.
func (p *provider) CreateCluster(cr *apistructs.ClusterCreateRequest) error {
	// Create cluster with bundle.
	// cluster-manager -> bundle -> cmdb -> eventBox -> scheduler (hook)
	//                               |                     |
	//                           *database*             *etcd*
	return p.bundle.CreateCluster(cr, http.Header{
		httputil.InternalHeader: []string{"cluster-manager"},
	})
}

// UpdateCluster Update cluster with bundle.
func (p *provider) UpdateCluster(cr *apistructs.ClusterUpdateRequest) error {
	if cr == nil {
		return errors.New("cluster update request is nil")
	}

	// Default header, must provider internal and org header.
	header := http.Header{
		httputil.InternalHeader: []string{"cluster-manager"},
		httputil.OrgHeader:      []string{strconv.Itoa(cr.OrgID)},
	}

	// If doesn't provider orgID, get first.
	if cr.OrgID == 0 {
		ci, err := p.bundle.GetCluster(cr.Name)
		if err != nil {
			return err
		}
		cr.OrgID = ci.OrgID
		header[httputil.OrgHeader] = []string{strconv.Itoa(ci.OrgID)}
	}

	// The process of update is same to create.
	return p.bundle.UpdateCluster(*cr, header)
}

// DeleteCluster Delete cluster with bundle.
func (p *provider) DeleteCluster(clusterName string) error {
	return p.bundle.DeleteCluster(clusterName, http.Header{
		httputil.InternalHeader: []string{"cluster-manager"},
	})
}
