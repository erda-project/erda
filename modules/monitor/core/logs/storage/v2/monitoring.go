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
		Name: "log_bytes_v2",
		Help: "the size of consumed log",
	},
		[]string{levelKey, srcKey, srcComponentTypeKey, srcComponentNameKey, srcClusterNameKey, srcOrgNameKey, srcProjectIDKey, srcProjectNameKey, srcApplicationIDKey, srcApplicationNameKey, srcWorkspaceKey})
)

const (
	platformKey        = "platform"
	componentKey       = "component"
	componentNameKey   = "component_name"
	componentTypeKey   = "component_type"
	orgIDKey           = "org_id"
	orgNameKey         = "org_name"
	clusterNameKey     = "cluster_name"
	projectIDKey       = "project_id"
	projectNameKey     = "project_name"
	applicationIDKey   = "application_id"
	applicationNameKey = "application_name"
	workspaceKey       = "workspace"
	levelKey           = "level"

	dicePrefix             = "dice_"
	diceComponentKey       = dicePrefix + componentKey
	diceOrgIDKey           = dicePrefix + orgIDKey
	diceOrgNameKey         = dicePrefix + orgNameKey
	diceClusterNameKey     = dicePrefix + clusterNameKey
	diceProjectIDKey       = dicePrefix + projectIDKey
	diceProjectNameKey     = dicePrefix + projectNameKey
	diceApplicationIDKey   = dicePrefix + applicationIDKey
	diceApplicationNameKey = dicePrefix + applicationNameKey
	diceWorkspaceKey       = dicePrefix + workspaceKey

	srcKey                = "src"
	srcPrefix             = "src_"
	srcComponentNameKey   = srcPrefix + componentNameKey
	srcComponentTypeKey   = srcPrefix + componentTypeKey
	srcOrgNameKey         = srcPrefix + orgNameKey
	srcClusterNameKey     = srcPrefix + clusterNameKey
	srcProjectIDKey       = srcPrefix + projectIDKey
	srcProjectNameKey     = srcPrefix + projectNameKey
	srcApplicationIDKey   = srcPrefix + applicationIDKey
	srcApplicationNameKey = srcPrefix + applicationNameKey
	srcWorkspaceKey       = srcPrefix + workspaceKey
)

// todo prometheus
// func countV2(log *pb.Log) {
// 	componentName := log.Tags[diceComponentKey]
// 	var componentType string
// 	if componentName != "" {
// 		componentType = platformKey
// 	}
// 	logBytesCounter.WithLabelValues(
// 		log.Tags[levelKey],
// 		log.Source,
// 		componentType,
// 		componentName,
// 		log.Tags[diceClusterNameKey],
// 		log.Tags[diceOrgNameKey],
// 		log.Tags[diceProjectIDKey],
// 		log.Tags[diceProjectNameKey],
// 		log.Tags[diceApplicationIDKey],
// 		log.Tags[diceApplicationNameKey],
// 		log.Tags[diceWorkspaceKey],
// 	).Add(float64(len(log.Content)))
// }

func init() {
	prometheus.MustRegister(logBytesCounter)
}
