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

// 租户标签
func OrgLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	// 获取 orgName 的两种途径
	// 1. 在 diceyml.meta (li.Label) 中的 labelconfig.ORG_KEY
	// 2. diceyml.deployments.selectors.org, 只要某个 service 存在即可，取第一个
	// 优先级 2 > 1
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
	// 从集群精细配置获取是否开启 org 调度
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

	// 从基本配置中获取是否开启 org 调度
	// 一般建议 org 是否开启放在精细配置中，基本配置中不配置 org，即默认不开启 org
	enableOrgScheduler := li.ExecutorConfig.EnableOrgLabelSchedule()
	// 未开启 org 调度
	if !enableOrgScheduler {
		r.UnLikePrefixs = append(r.UnLikePrefixs, labelconfig.ORG_VALUE_PREFIX)
		r2.HasOrg = false
		return
	}

	r.ExclusiveLikes = append(r.ExclusiveLikes, labelconfig.ORG_VALUE_PREFIX+orgName)
	r2.HasOrg = true
	r2.Org = orgName
}
