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

package externalprovider

import (
	"github.com/erda-project/erda/internal/apps/msp/instance/db"
	"github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers"
)

const (
	MseNacosHost = "MS_NACOS_HOST"
	MseNacosPort = "MS_NACOS_PORT"
)

// TODO it will support like "TSE" "HSE"
var matches = []string{handlers.ResourceMSENacos}

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	for _, res := range matches {
		if res == tmc.Engine {
			return true
		}
	}
	return false
}

func (p *provider) CheckIfHasCustomConfig(clusterConfig map[string]string) (map[string]string, bool) {
	nacosHost, ok := clusterConfig[MseNacosHost]
	if !ok {
		return nil, false
	}

	nacosPort, ok := clusterConfig[MseNacosPort]
	if !ok {
		return nil, false
	}

	return map[string]string{
		"NACOS_HOST":     nacosHost,
		"NACOS_PORT":     nacosPort,
		"NACOS_ADDRESS":  nacosHost + ":" + nacosPort,
		"NACOS_USER":     "nacos",
		"NACOS_PASSWORD": "nacos",
	}, true
}
