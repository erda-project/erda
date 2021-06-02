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

package api

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"

	"github.com/erda-project/erda/bundle"
)

var (
	rc  *reqCli
	bdl = bundle.New(bundle.WithCMDB(), bundle.WithOps())
)

func init() {
	rc = &reqCli{
		vendors: make(map[string]map[string]CloudVendor),
	}
}

type reqCli struct {
	// map[orgId][vendor_name]vendor. because one org may use different vendor
	vendors map[string]map[string]CloudVendor
}

func RegisterVendor(orgId string, vendor CloudVendor) {
	if _, ok := rc.vendors[orgId]; !ok {
		rc.vendors[orgId] = make(map[string]CloudVendor)
	}
	switch vendor.(type) {
	case *aliyun:
		rc.vendors[orgId]["aliyun"] = vendor
	default:
	}
}

// DoReqToAliyun .
func (r *reqCli) DoReqToAliyun(orgId string, request requests.AcsRequest, response responses.AcsResponse) error {
	return r.DoReqToVendor(orgId, request, response, "aliyun")
}

func (r *reqCli) DoReqToVendor(orgId string, request interface{}, response interface{}, vendor string) error {
	org, ok := r.vendors[orgId]
	if !ok {
		return fmt.Errorf("can not find org. orgId=%s", orgId)
	}
	v, ok := org[vendor]
	if !ok {
		return nil
	}
	return v.DoReq(request, response)
}
