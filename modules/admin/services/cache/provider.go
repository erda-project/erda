// Copyright (c) 2022 Terminus, Inc.
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

package cache

import (
	"reflect"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/cache"
)

var (
	name      = "easy-memory-cache"
	cacheType = reflect.TypeOf((*Cache)(nil))
	spec      = servicehub.Spec{
		Services:    []string{name, name + "-client"},
		Types:       []reflect.Type{cacheType},
		Description: "easy-memory-cache",
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	}
)

func init() {
	servicehub.Register(name, &spec)
}

// +provider
type provider struct{}

// Init .
func (p *provider) Init(ctx servicehub.Context) error {
	logrus.WithField("provider", name).Infoln("Init")
	if c == nil {
		c = new(Cache)
	}
	bdl := bundle.New(bundle.WithCoreServices())
	orgID2Org = cache.New("dop-org-id-for-org", time.Hour*24, func(i interface{}) (interface{}, bool) {
		orgDTO, err := bdl.GetOrg(i.(string))
		if err != nil {
			return nil, false
		}
		return orgDTO, true
	})
	proID2Org = cache.New("dop-project-id-for-org", time.Minute*30, func(i interface{}) (interface{}, bool) {
		projectID, err := strconv.ParseUint(i.(string), 10, 32)
		if err != nil {
			return nil, false
		}
		projectDTO, err := bdl.GetProject(projectID)
		if err != nil {
			return nil, false
		}
		orgDTO, ok := c.GetOrgByOrgID(strconv.FormatUint(projectDTO.OrgID, 10))
		if !ok {
			return nil, false
		}
		return orgDTO, true
	})
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	logrus.WithField("provider", name).Infoln("Provide")
	if ctx.Service() == name+"-client" || ctx.Type() == cacheType {
		if c == nil {
			c = new(Cache)
		}
		return c
	}
	return p
}
