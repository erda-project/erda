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

package storage

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	logBytesCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "log_bytes",
		Help: "the size of consumed log",
	},
		[]string{levelKey, srcKey, srcComponentTypeKey, srcComponentNameKey, srcClusterNameKey, srcOrgNameKey, srcProjectIDKey, srcProjectNameKey, srcApplicationIDKey, srcApplicationNameKey, srcWorkspaceKey})
)

func init() {
	prometheus.MustRegister(logBytesCounter)
}
