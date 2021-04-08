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

package labelpipeline

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
)

// Tenant label
func OrgLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	// Two ways to get orgName
	// 1. Labelconfig.ORG_KEY in diceyml.meta (li.Label)
	// 2. diceyml.deployments.selectors.org, as long as a certain service exists, take the first one
	// Priority 2> 1
	orgName := li.Label[labelconfig.ORG_KEY]
	for _, selectors := range li.Selectors {
		if v, ok := selectors["org"]; ok && len(v.Values) > 0 {
			orgName = v.Values[0]
			break
		}
	}
	if orgName == "" {
		r2.HasOrg = false
		logrus.Infof("obj(%s) have no orgName", li.ObjName)
		return
	}
	// Get whether to enable org scheduling from the cluster fine configuration
	if li.OptionsPlus != nil {
		for _, orgCfg := range li.OptionsPlus.Orgs {
			if orgCfg.Name == orgName {
				if v, orgOK := orgCfg.Options[labelconfig.ENABLE_ORG]; orgOK && v == "true" {
					r.ExclusiveLikes = append(r.ExclusiveLikes, labelconfig.ORG_VALUE_PREFIX+orgName)
					r2.HasOrg = true
					r2.Org = orgName
					logrus.Infof("obj(%s) got refined org configuration", li.ObjName)
					return
				}
				break
			}
		}
	}

	// Get whether to enable org scheduling from the basic configuration
	// It is generally recommended that whether org is enabled or not is placed in the fine configuration, and org is not configured in the basic configuration, that is, org is not enabled by default
	enableOrgScheduler := li.ExecutorConfig.EnableOrgLabelSchedule()
	// Org scheduling is not turned on
	if !enableOrgScheduler {
		r.UnLikePrefixs = append(r.UnLikePrefixs, labelconfig.ORG_VALUE_PREFIX)
		r2.HasOrg = false
		return
	}

	r.ExclusiveLikes = append(r.ExclusiveLikes, labelconfig.ORG_VALUE_PREFIX+orgName)
	r2.HasOrg = true
	r2.Org = orgName
}
