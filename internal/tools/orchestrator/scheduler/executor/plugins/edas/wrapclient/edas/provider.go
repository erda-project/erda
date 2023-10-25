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

package edas

import (
	api "github.com/aliyun/alibaba-cloud-sdk-go/services/edas"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/types"
)

var notFound = "not found"

type wrapEDAS struct {
	l *logrus.Entry

	client          *api.Client
	addr            string
	clusterID       string
	regionID        string
	logicalRegionID string
	unLimitCPU      string
}

// refactor, source: edas/edas.go
// PR: https://github.com/erda-project/erda/pull/6102

type Interface interface {
	GetAppID(name string) (string, error)
	GetAppDeployment(appName string) (*appsv1.Deployment, error)
	QueryAppStatus(appName string) (types.AppStatus, error)
	InsertK8sApp(spec *types.ServiceSpec) (string, error)
	DeployApp(appID string, spec *types.ServiceSpec) error
	DeleteAppByName(appName string) error
	ScaleApp(appID string, replica int) error
	AbortChangeOrder(changeOrderID string) error
	LoopTerminationStatus(orderID string) (types.ChangeOrderStatus, error)
	ListRecentChangeOrderInfo(appID string) (*api.ChangeOrderList, error)
}

func New(l *logrus.Entry, client *api.Client, addr, clusterId, regionID, logicalRegionID, unLimitCPU string) Interface {
	return &wrapEDAS{
		l:               l.WithField("wrap-client", "edas"),
		client:          client,
		addr:            addr,
		clusterID:       clusterId,
		regionID:        regionID,
		logicalRegionID: logicalRegionID,
		unLimitCPU:      unLimitCPU,
	}
}
