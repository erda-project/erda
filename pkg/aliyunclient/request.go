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

package aliyunclient

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/pkg/errors"
)

type RemoteActionRequest struct {
	ClusterName          string
	Product              string
	Version              string
	ActionName           string
	LocationServiceCode  string
	LocationEndpointType string
	QueryParams          map[string]string
	Headers              map[string]string
	FormParams           map[string]string
}

func NewRemoteActionRequest(acsReq requests.AcsRequest) (*RemoteActionRequest, error) {
	req := &RemoteActionRequest{
		Product:              acsReq.GetProduct(),
		Version:              acsReq.GetVersion(),
		ActionName:           acsReq.GetActionName(),
		LocationServiceCode:  acsReq.GetLocationServiceCode(),
		LocationEndpointType: acsReq.GetLocationEndpointType(),
	}
	err := requests.InitParams(acsReq)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.QueryParams = acsReq.GetQueryParams()
	req.Headers = acsReq.GetHeaders()
	req.FormParams = acsReq.GetFormParams()
	return req, nil
}
