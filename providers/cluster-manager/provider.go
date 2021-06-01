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
	"reflect"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/clientgo"
)

// Interface .
type Interface interface {
	GetClusterClientSet(clusterName string) (*clientgo.ClientSet, error)
	GetCluster(clusterName string) (*apistructs.ClusterInfo, error)
	CreateCluster(cr *apistructs.ClusterCreateRequest) error
	UpdateCluster(cr *apistructs.ClusterUpdateRequest) error
	DeleteCluster(clusterName string) error
}

type define struct{}

func (d *define) Services() []string { return []string{"cluster-manager"} }
func (d *define) Types() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Interface)(nil)).Elem(),
	}
}
func (d *define) Description() string { return "cluster-manager" }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type provider struct {
	bundle *bundle.Bundle
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bundle = bundle.New(bundle.WithCMDB())

	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return p
}

func init() {
	servicehub.RegisterProvider("cluster-manager", &define{})
}
