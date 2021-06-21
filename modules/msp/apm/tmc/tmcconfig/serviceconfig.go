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

package tmcconfig

import "github.com/erda-project/erda/pkg/envconf"

var Conf Monitor

type Monitor struct {
	MonitorAddr                 string `env:"MONITOR_ADDR" default:"monitor.default.svc.cluster.local:7096"`
	MonitorServiceMetricApiPath string `env:"MONITOR_METRIC_PATH" default:"/api/metrics"`
}

func Load() {
	envconf.MustLoad(&Conf)
}

func init() {
	Load()
}
