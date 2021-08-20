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

package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"golang.org/x/text/message"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (e *Endpoints) ListLabels(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	i18n := ctx.Value("i18nPrinter").(*message.Printer)
	return mkResponse(apistructs.ListLabelsResponse{
		Header: apistructs.Header{Success: true},
		Data: []apistructs.ListLabelsData{
			{
				Name:       i18n.Sprintf("locked"),
				Label:      "locked",
				Desc:       "",
				Group:      "others",
				GroupName:  i18n.Sprintf("others"),
				GroupLevel: 5,
			},
			{
				Name:       i18n.Sprintf("platform"),
				Label:      "platform",
				Desc:       "",
				Group:      "platform",
				GroupName:  i18n.Sprintf("platform"),
				GroupLevel: 1,
			},
			{
				Name:       i18n.Sprintf("pack-job"),
				Label:      "pack-job",
				Desc:       "",
				Group:      "job",
				GroupName:  i18n.Sprintf("job"),
				GroupLevel: 4,
			},
			{
				Name:       i18n.Sprintf("bigdata-job"),
				Label:      "bigdata-job",
				Desc:       "",
				Group:      "job",
				GroupName:  i18n.Sprintf("job"),
				GroupLevel: 4,
			},
			{
				Name:       i18n.Sprintf("job"),
				Label:      "job",
				Desc:       "",
				Group:      "job",
				GroupName:  i18n.Sprintf("job"),
				GroupLevel: 4,
			},
			{
				Name:       i18n.Sprintf("stateful-service"),
				Label:      "stateful-service",
				Desc:       "",
				Group:      "service",
				GroupName:  i18n.Sprintf("service"),
				GroupLevel: 3,
			},
			{
				Name:       i18n.Sprintf("stateless-service"),
				Label:      "stateless-service",
				Desc:       "",
				Group:      "service",
				GroupName:  i18n.Sprintf("service"),
				GroupLevel: 3,
			},
			{
				Name:       i18n.Sprintf("location-cluster-service"),
				Label:      "location-cluster-service",
				Desc:       "",
				Group:      "service",
				GroupName:  i18n.Sprintf("service"),
				GroupLevel: 3,
			},
			{
				Name:       i18n.Sprintf("workspace-dev"),
				Label:      "workspace-dev",
				Desc:       "",
				Group:      "env",
				GroupName:  i18n.Sprintf("env"),
				GroupLevel: 2,
			},
			{
				Name:       i18n.Sprintf("workspace-test"),
				Label:      "workspace-test",
				Desc:       "",
				Group:      "env",
				GroupName:  i18n.Sprintf("env"),
				GroupLevel: 2,
			},
			{
				Name:       i18n.Sprintf("workspace-staging"),
				Label:      "workspace-staging",
				Desc:       "",
				Group:      "env",
				GroupName:  i18n.Sprintf("env"),
				GroupLevel: 2,
			},
			{
				Name:       i18n.Sprintf("workspace-prod"),
				Label:      "workspace-prod",
				Desc:       "",
				Group:      "env",
				GroupName:  i18n.Sprintf("env"),
				GroupLevel: 2,
			},
			{
				Name:       i18n.Sprintf("org-"),
				Label:      "org-",
				Desc:       "",
				IsPrefix:   true,
				Group:      "others",
				GroupName:  i18n.Sprintf("others"),
				GroupLevel: 5,
			},
			{
				Name:       i18n.Sprintf("location-"),
				Label:      "location-",
				Desc:       "",
				IsPrefix:   true,
				Group:      "others",
				GroupName:  i18n.Sprintf("others"),
				GroupLevel: 5,
			},
			{
				Name:       i18n.Sprintf("topology-zone"),
				Label:      "topology-zone",
				Desc:       "",
				Group:      "others",
				GroupName:  i18n.Sprintf("others"),
				GroupLevel: 5,
				WithValue:  true,
			},
		},
	})
}

func (e *Endpoints) UpdateLabels(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("error happened, error:%v", err)
			resp, err = mkResponse(apistructs.CloudClusterResponse{
				Header: apistructs.Header{
					Success: false,
					Error:   apistructs.ErrorResponse{Msg: err.Error()},
				},
			})
		}
	}()

	var req apistructs.UpdateLabelsRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("failed to unmarshal to apistructs.UpdateLabelsRequest: %v", err)
		return
	}

	i, resp := e.GetIdentity(r)
	if resp != nil {
		err = fmt.Errorf("failed to get User-ID or Org-ID from request header")
		return
	}

	// permission check
	err = e.PermissionCheck(i.UserID, i.OrgID, "", apistructs.UpdateAction)
	if err != nil {
		return
	}

	recordID, err := e.labels.UpdateLabels(req, i.UserID)
	if err != nil {
		err = fmt.Errorf("failed to updatelabels: %v", err)
		return
	}

	return mkResponse(apistructs.UpdateLabelsResponse{
		Header: apistructs.Header{
			Success: true,
		},
		Data: apistructs.UpdateLabelsData{
			RecordID: recordID,
		},
	})
}
