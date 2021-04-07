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
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/strutil"
)

// Workspace 标签(即 env 标签，区分 dev, test, staging, prod)
func WorkspaceLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	envName_ := li.Label[labelconfig.WORKSPACE_KEY]
	envName := strings.ToLower(envName_)
	for _, selectors := range li.Selectors {
		if v, ok := selectors["workspace"]; ok && len(v.Values) > 0 {
			envName = v.Values[0]
			break
		}
	}
	if envName == "" {
		// 未打 env 标的应用，不应该被调度到任意一个已打了该标签的节点上
		r.UnLikePrefixs = append(r.UnLikePrefixs, labelconfig.WORKSPACE_VALUE_PREFIX)
		r2.HasWorkSpace = false
		logrus.Infof("obj(%s) have no workspace label", li.ObjName)
		return
	}

	var (
		workspaceopt *executortypes.WorkspaceOpt
		orgopt       *executortypes.OrgOpt
	)
	// 从集群精细配置获取是否开启 workspace 调度，workspace 的配置在 org 之下
	orgName, orgOK := li.Label[labelconfig.ORG_KEY]
	for _, selectors := range li.Selectors {
		if v, ok := selectors["org"]; ok && len(v.Values) > 0 {
			orgName = v.Values[0]
			orgOK = true
			break
		}
	}
	// 未打 org 标的应用，在精细配置中无法找到是否开启 workspace 调度的配置
	// 去集群基本配置中读取是否开启 workspace 调度
	if !orgOK || li.OptionsPlus == nil {
		goto basic
	}

	orgopt = li.ExecutorConfig.OrgOpt(orgName)
	if orgopt == nil {
		goto basic
	}

	workspaceopt = orgopt.WorkspaceOpt(envName)
	if workspaceopt == nil {
		goto basic
	}

	if workspaceopt.EnableWorkspaceLabelSchedule() {
		if fixJobDest(r, r2, li, orgName, envName, li.ExecutorConfig, false) {
			return
		}
		r.ExclusiveLikes = append(r.ExclusiveLikes, labelconfig.WORKSPACE_VALUE_PREFIX+envName)
		r2.HasWorkSpace = true
		r2.WorkSpaces = []string{envName}
		// We have already set up the `workspace' label, and we
		// don't have to perform the `basic' part of the process
		// below, so just return here.
		return
	}

basic:
	// 从基本配置获取是否开启 workspace 调度
	enableEnvScheduler := li.ExecutorConfig.EnableWorkspaceLabelSchedule()
	// 未开启 workspace 调度
	if !enableEnvScheduler {
		// 兼容老的 WORKSPACETAGS 标签
		if tag, tagOK := li.ExecutorConfig.WORKSPACETAGSForCompatibility(); tagOK {
			tags := strings.Split(tag, ",")
			sort.Strings(tags)
			idx := sort.SearchStrings(tags, envName)
			if idx < len(tags) && tags[idx] == envName {
				r.ExclusiveLikes = append(r.ExclusiveLikes, labelconfig.WORKSPACE_VALUE_PREFIX+envName)
				r2.HasWorkSpace = true
				r2.WorkSpaces = []string{envName}
			} else {
				r.UnLikePrefixs = append(r.UnLikePrefixs, labelconfig.WORKSPACE_VALUE_PREFIX)
				r2.HasWorkSpace = false
			}
			return
		}
		r.UnLikePrefixs = append(r.UnLikePrefixs, labelconfig.WORKSPACE_VALUE_PREFIX)
		r2.HasWorkSpace = false
		return
	}

	if fixJobDest(r, r2, li, orgName, envName, li.ExecutorConfig, true) {
		return
	}
	r.ExclusiveLikes = append(r.ExclusiveLikes, labelconfig.WORKSPACE_VALUE_PREFIX+envName)
	r2.HasWorkSpace = true
	r2.WorkSpaces = []string{envName}
}

// fixJobDest 对 Staging 和 Prod 的 job 可以额外配置目的工作区，相应配置如下:
// "STAGING_JOB_DEST":"dev"
// "PROD_JOB_DEST":"dev,test"
func fixJobDest(r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo, org, workspace string, options *executortypes.ExecutorWholeConfigs, basicconfig bool) bool {
	if (li.ExecutorKind != labelconfig.EXECUTOR_METRONOME) && (li.ExecutorKind != labelconfig.EXECUTOR_K8SJOB) {
		return false
	}
	// The caller guarantees that 'opt' ​​is non-nil
	var opt interface {
		StagingJobAvailDest() ([]string, bool)
		ProdJobAvailDest() ([]string, bool)
	} = options
	if !basicconfig {
		opt = options.OrgOpt(org).WorkspaceOpt(workspace)
	}
	var (
		dests  []string
		destOK bool
	)
	switch workspace {
	case labelconfig.WORKSPACE_STAGING:
		dests, destOK = opt.StagingJobAvailDest()
	case labelconfig.WORKSPACE_PROD:
		dests, destOK = opt.ProdJobAvailDest()
	default:
		return false
	}
	if !destOK {
		// Default behavior when 'STAGING_JOB_DEST' or 'PROD_JOB_DEST' is not set, but is this really appropriate?
		r.InclusiveLikes = append(r.InclusiveLikes,
			labelconfig.WORKSPACE_VALUE_PREFIX+labelconfig.WORKSPACE_DEV)
		r.InclusiveLikes = append(r.InclusiveLikes,
			labelconfig.WORKSPACE_VALUE_PREFIX+labelconfig.WORKSPACE_TEST)
		r2.HasWorkSpace = true
		// k8s 集群不配置默认值，按实际传递的 workspace 进行调度
		r2.WorkSpaces = append(r2.WorkSpaces, workspace)
		return true
	}
	r.InclusiveLikes = append(r.InclusiveLikes, strutil.Map(dests, func(s string) string {
		return strutil.Concat(labelconfig.WORKSPACE_VALUE_PREFIX, s)
	})...)
	r2.HasWorkSpace = true
	r2.WorkSpaces = append(r2.WorkSpaces, dests...)
	return true
}
