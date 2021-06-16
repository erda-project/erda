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

// Package bundle 见 bundle.go
package bundle

import (
	"fmt"
	"os"
	"strings"

	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/discover"
)

type urls map[string]string

func (u urls) Put(k, v string) {
	// v 为空说明没有在环境变量开启对应的客户端.
	if v == "" {
		return
	}
	if !validateURL(v) {
		panic(fmt.Sprintf("invalid env: \"%s\"", v))
	}
	if _, ok := u[k]; ok {
		panic(fmt.Sprintf("duplicate env: \"%s\"", k))
	}
	u[k] = v
}

func (u urls) PutAllAvailable() {
	for _, kv := range os.Environ() {
		ss := strings.SplitN(kv, "=", 2)
		if len(ss) < 2 {
			continue
		}
		u.Put(ss[0], ss[1])
	}
}

func (u urls) EventBox() (string, error) {
	return u.getURL(discover.EnvEventBox, discover.SvcEventBox)
}

func (u urls) CMDB() (string, error) {
	return u.getURL(discover.EnvCMDB, discover.SvcCMDB)
}

func (u urls) Scheduler() (string, error) {
	return u.getURL(discover.EnvScheduler, discover.SvcScheduler)
}

func (u urls) DiceHub() (string, error) {
	return u.getURL(discover.EnvDiceHub, discover.SvcDiceHub)
}

func (u urls) Soldier() (string, error) {
	return u.getURL(discover.EnvSoldier, discover.SvcSoldier)
}

func (u urls) Orchestrator() (string, error) {
	return u.getURL(discover.EnvOrchestrator, discover.SvcOrchestrator)
}

func (u urls) AddOnPlatform() (string, error) {
	return u.getURL(discover.EnvAddOnPlatform, discover.SvcAddOnPlatform)
}

func (u urls) Gittar() (string, error) {
	return u.getURL(discover.EnvGittar, discover.SvcGittar)
}

func (u urls) GittarAdaptor() (string, error) {
	return u.getURL(discover.EnvGittarAdaptor, discover.SvcGittarAdaptor)
}

func (u urls) Collector() (string, error) {
	return u.getURL(discover.EnvCollector, discover.SvcCollector)
}

func (u urls) Monitor() (string, error) {
	return u.getURL(discover.EnvMonitor, discover.SvcMonitor)
}

func (u urls) Pipeline() (string, error) {
	return u.getURL(discover.EnvPipeline, discover.SvcPipeline)
}

func (u urls) Hepa() (string, error) {
	return u.getURL(discover.EnvHepa, discover.SvcHepa)
}

func (u urls) TMC() (string, error) {
	return u.getURL(discover.EnvTMC, discover.SvcTMC)
}

func (u urls) Ops() (string, error) {
	return u.getURL(discover.EnvOps, discover.SvcOps)
}

func (u urls) Openapi() (string, error) {
	return u.getURL(discover.EnvOpenapi, discover.SvcOpenapi)
}

func (u urls) KMS() (string, error) {
	return u.getURL(discover.EnvKMS, discover.SvcKMS)
}

func (u urls) QA() (string, error) {
	return u.getURL(discover.EnvQA, discover.SvcQA)
}

func (u urls) APIM() (string, error) {
	return u.getURL(discover.EnvAPIM, discover.SvcAPIM)
}

func (u urls) DOP() (string, error) {
	return u.getURL(discover.EnvDOP, discover.SvcDOP)
}

func (u urls) ClusterManager() (string, error) {
	return u.getURL(discover.EnvClusterManager, discover.SvcClusterManager)
}

func (u urls) ECP() (string, error) {
	return u.getURL(discover.EnvECP, discover.SvcECP)
}

func (u urls) getURL(k, srvName string) (string, error) {
	v, ok := u[k]
	if ok {
		return v, nil
	}
	if srvName != "" {
		return discover.GetEndpoint(srvName)
	}
	return v, apierrors.ErrUnavailableClient.InvalidState(k)
}

// TODO
func validateURL(url string) bool {
	return true
}
