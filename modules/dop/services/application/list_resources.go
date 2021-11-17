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

package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	dashboardPb "github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

func (app *Application) ApplicationsResources(ctx context.Context, req *apistructs.ApplicationsResourcesRequest)(
	*apistructs.ApplicationsResourcesResponse, *errorresp.APIError) {
	l := logrus.WithField("func", "*Application.ApplicationsResources")
	if req == nil {
		err := errors.New("request can not be nil")
		l.WithError(err).Errorln("the req is nil")
		return nil, apierrors.ErrApplicationsResources.InvalidParameter(err)
	}

	var data apistructs.ApplicationsResourcesResponse

	// query applications list from core-services
	// todo: pageSize
	apps, err := app.bdl.GetAppsByProject(req.ProjectID, req.OrgID, req.UserID)
	if err != nil {
		l.WithError(err).Errorln("failed to GetAppsByProject")
		return nil, apierrors.ErrApplicationsResources.InternalError(err)
	}
	if apps.Total == 0 || len(apps.List) == 0 {
		return &data, nil
	}

	// query runtimes list from orchestrator
	var applicationsIDs []uint64
	for _, application := range apps.List {
		applicationsIDs = append(applicationsIDs, application.ID)
	}
	runtimesM, err := app.bdl.GetApplicationsRuntimes(req.OrgID, req.UserID, applicationsIDs)
	if err != nil {
		l.WithError(err).Errorln("failed to GetApplicationsRuntimes")
		return nil, apierrors.ErrApplicationsResources.InternalError(err)
	}

	var (
		// servicesGroupsOnClusters {serviceGroupID:clusterName}
		servicesGroupsOnClusters = make(map[string]string)
		// servicesGroupsOnApps {serviceGroupID:applicationID}
		servicesGroupsOnApps = make(map[string]uint64)
		// clustersServiceGroups {clusterName:serviceGroupID}
		clustersServiceGroups = make(map[string][]string)
	)
	for applicationID, runtimes := range runtimesM {
		for _, runtime := range runtimes {
			servicesGroupsOnClusters[runtime.ServiceGroupName] = runtime.ClusterName
			servicesGroupsOnApps[runtime.ServiceGroupName] = applicationID
			clustersServiceGroups[runtime.ClusterName] = append(clustersServiceGroups[runtime.ClusterName], runtime.ServiceGroupName)
		}
	}
	for cluster, labels := range clustersServiceGroups {
		app.cmp.GetPodsByLabels(ctx, &dashboardPb.GetPodsByLabelsRequest{
			Cluster: cluster,
			Labels:  []string{fmt.Sprintf("servicegroup-id in (%s)", strings.Join(labels, ","))},
		})
	}
}
