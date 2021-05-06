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
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) ClusterPreview(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	i18nPrinter := ctx.Value("i18nPrinter").(*message.Printer)
	var req apistructs.CloudClusterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("failed to unmarshal to apistructs.CloudClusterRequest: %v", err)
		return mkResponse(apistructs.ClusterPreviewResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	logrus.Debugf("cloud-cluster request: %v", req)
	userid := r.Header.Get("User-ID")
	if userid == "" {
		errstr := fmt.Sprintf("failed to get user-id in http header")
		return mkResponse(apistructs.ClusterPreviewResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	data, err := e.clusters.ClusterPreview(req)
	if err != nil {
		errstr := fmt.Sprintf("failed to preview cluster info: %v", err)
		return mkResponse(apistructs.ClusterPreviewResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}

	for i, resource := range data {
		rsc := i18nPrinter.Sprintf(string(resource.Resource))
		data[i].Resource = apistructs.ClusterResourceType(rsc)
		for j, profile := range resource.ResourceProfile {
			data[i].ResourceProfile[j] = i18nPrinter.Sprintf(profile)
		}
	}

	return mkResponse(apistructs.ClusterPreviewResponse{
		Header: apistructs.Header{Success: true},
		Data:   data,
	})
}
