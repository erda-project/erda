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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	g "github.com/erda-project/erda/modules/extensions/cloud/aliyun/metrics/cloudcat/globals"
)

var (
	ErrEmptyResults = errors.New("empty results")
)

// todo need as a requests pool
func GetDescribeMetricLast(orgId, namespace, metricName string) (dataPoints []string, err error) {
	req := cms.CreateDescribeMetricLastRequest()
	req.Namespace = namespace
	req.MetricName = metricName
	dp := []string{}
	if err := recursiveGetDescribeMetricLast(orgId, req, &dp); err != nil {
		return nil, err
	}
	return dp, nil
}

func recursiveGetDescribeMetricLast(orgId string, req *cms.DescribeMetricLastRequest, datapoint *[]string) error {
	resp := cms.CreateDescribeMetricLastResponse()
	err := rc.DoReqToAliyun(orgId, req, resp)
	if err != nil {
		return err
	}
	if resp.Datapoints == "" || resp.Datapoints == "[]" {
		return ErrEmptyResults
	}
	*datapoint = append(*datapoint, resp.Datapoints)
	if resp.NextToken != "" {
		req.NextToken = resp.NextToken
		return recursiveGetDescribeMetricLast(orgId, req, datapoint)
	}
	return nil
}

func ListMetricMeta(orgId, namespace string) (res []cms.Resource, err error) {
	// get total
	req := cms.CreateDescribeMetricMetaListRequest()
	req.Namespace = namespace
	req.PageSize = requests.NewInteger(1)
	resp := cms.CreateDescribeMetricMetaListResponse()
	err = rc.DoReqToAliyun(orgId, req, resp)
	if err != nil || !resp.Success {
		return nil, err
	}
	if len(resp.Resources.Resource) == 0 {
		return nil, ErrEmptyResults
	}

	// fetch
	total, err := strconv.Atoi(resp.TotalCount)
	if err != nil || !resp.Success {
		return nil, err
	}
	req.PageSize = requests.NewInteger(total)
	err = rc.DoReqToAliyun(orgId, req, resp)
	return resp.Resources.Resource, nil
}

func ListProjectMeta(orgId string, products []string) (res []cms.Resource, err error) {
	req := cms.CreateDescribeProjectMetaRequest()
	req.PageSize = requests.NewInteger(1)
	resp := cms.CreateDescribeProjectMetaResponse()
	err = rc.DoReqToAliyun(orgId, req, resp)
	if err != nil || !resp.Success {
		return nil, err
	}
	if len(resp.Resources.Resource) == 0 {
		return nil, ErrEmptyResults
	}

	// fetch
	total, err := strconv.Atoi(resp.Total)
	if err != nil {
		return nil, err
	}
	req.PageSize = requests.NewInteger(total)

	if products != nil && len(products) != 0 {
		res = make([]cms.Resource, 0)
		for _, p := range products {
			req.Labels = createLabels(p)
			err = rc.DoReqToAliyun(orgId, req, resp)
			if err != nil {
				return nil, err
			}
			res = append(res, resp.Resources.Resource...)
		}
		return res, nil
	} else {
		err = rc.DoReqToAliyun(orgId, req, resp)
		return resp.Resources.Resource, err
	}
}

func createLabels(product string) string {
	rv, _ := json.Marshal([]map[string]string{{"name": "product", "value": product}})
	return string(rv)
}

func ListOrgInfos() (res []OrgInfo, err error) {
	var list []apistructs.OrgDTO
	if g.Cfg.OrgIds != "" {
		for _, oid := range strings.Split(g.Cfg.OrgIds, ",") {
			org, err := bdl.GetOrg(oid)
			if err != nil {
				g.Log.Infof("failed to get org by orgid=%s", oid)
				continue
			}
			list = append(list, *org)
		}
	} else {
		list, err = ListAllOrgs()
		if err != nil {
			return nil, err
		}
	}

	for _, org := range list {
		orgId := strconv.Itoa(int(org.ID))
		account, err := GetAccountByOrgId(orgId)
		if err != nil {
			g.Log.Infof("failt to get aliyun account with orgId %s", orgId)
			continue
		}

		res = append(res, OrgInfo{orgId, org.Name, account.AccessKeyID, account.AccessSecret})
	}
	return
}

func ListAllOrgs() (res []apistructs.OrgDTO, err error) {
	req := &apistructs.OrgSearchRequest{
		Q:        "",
		PageNo:   1,
		PageSize: 1,
	}
	resp, err := bdl.ListDopOrgs(req)
	if err != nil {
		return nil, errors.Wrap(err, "ListOrgs err")
	}
	req.PageSize = resp.Total
	resp, err = bdl.ListDopOrgs(req)
	if err != nil {
		return nil, errors.Wrap(err, "ListOrgs err")
	}

	if len(resp.List) == 0 {
		return nil, errors.Errorf("no org")
	}
	return resp.List, nil
}

func GetAccountByOrgId(orgId string) (ac *apistructs.CloudAccount, err error) {
	account, err := bdl.GetOrgAccount(orgId, "aliyun")
	if account == nil || err != nil || account.AccessKeyID == "" {
		return nil, fmt.Errorf("orgId %s don't have aliyun account", orgId)

	}
	return account, nil
}
