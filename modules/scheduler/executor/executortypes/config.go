package executortypes

import (
	"strconv"

	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/pkg/strutil"
)

// EnableLabelSchedule 是否开启了标签调度
func (c *ExecutorWholeConfigs) EnableLabelSchedule() bool {
	if v, ok := c.BasicConfig["ENABLETAG"]; ok && len(v) > 0 {
		if enable, err := strconv.ParseBool(v); err == nil {
			return enable
		}
	}
	return false
}

// ProjectIDForCompatibility 调度的时候隔离 project 所用的配置，不再使用
// DEPRECATED
func (c *ExecutorWholeConfigs) ProjectIDForCompatibility(projectID string) bool {
	_, exist := c.BasicConfig[projectID]
	return exist
}

// EnableOrgLabelSchedule 是否开启了 org 级别的标签调度
func (c *ExecutorWholeConfigs) EnableOrgLabelSchedule() bool {
	if !c.EnableLabelSchedule() {
		return false
	}
	if enable, ok := c.BasicConfig["ENABLE_ORG"]; ok && strutil.ToLower(enable) == "true" {
		return true
	}
	return false
}

// EnableWorkspaceLabelSchedule 是否开启 workspace 级别的标签调度
func (c *ExecutorWholeConfigs) EnableWorkspaceLabelSchedule() bool {
	if !c.EnableLabelSchedule() {
		return false
	}
	if enable, ok := c.BasicConfig["ENABLE_WORKSPACE"]; ok && strutil.ToLower(enable) == "true" {
		return true
	}
	return false
}

// WORKSPACETAGSForCompatibility 兼容老的 WORKSPACETAGS 标签
// DEPRECATED
func (c *ExecutorWholeConfigs) WORKSPACETAGSForCompatibility() (string, bool) {
	v, ok := c.BasicConfig["WORKSPACETAGS"]
	return v, ok
}

// StagingJobAvailDest staging 环境的 JOB 可以运行的环境
func (c *ExecutorWholeConfigs) StagingJobAvailDest() ([]string, bool) {
	dests, ok := c.BasicConfig["STAGING_JOB_DEST"]
	if !ok {
		return nil, ok
	}
	return strutil.TrimSlice(strutil.Split(dests, ",", true)), true
}

// ProdJobAvailDest prod 环境的 job 可以运行的环境
func (c *ExecutorWholeConfigs) ProdJobAvailDest() ([]string, bool) {
	dests, ok := c.BasicConfig["PROD_JOB_DEST"]
	if !ok {
		return nil, ok
	}
	return strutil.TrimSlice(strutil.Split(dests, ",", true)), true
}

// OrgOpt org 级别的配置
type OrgOpt conf.Org

// WorkspaceOpt workspace 级别的配置
type WorkspaceOpt conf.Workspace

// OrgOpt 从所有配置中取出 `org' 的配置
func (c *ExecutorWholeConfigs) OrgOpt(org string) *OrgOpt {
	for _, orgopt := range c.PlusConfigs.Orgs {
		if orgopt.Name == org {
			opt := OrgOpt(orgopt)
			return &opt
		}
	}
	return nil
}

// WorkspaceOpt 从 org 配置中取出 `workspace' 的配置
func (c *OrgOpt) WorkspaceOpt(workspace string) *WorkspaceOpt {
	for _, workspaceopt := range c.Workspaces {
		if workspaceopt.Name == workspace {
			opt := WorkspaceOpt(workspaceopt)
			return &opt
		}
	}
	return nil
}

// EnableWorkspaceLabelSchedule 是否开启了 workspace 标签调度
func (c *WorkspaceOpt) EnableWorkspaceLabelSchedule() bool {
	if enable, ok := c.Options["ENABLE_WORKSPACE"]; ok && strutil.ToLower(enable) == "true" {
		return true
	}
	return false
}

// StagingJobAvailDest staging job 可运行的环境
func (c *WorkspaceOpt) StagingJobAvailDest() ([]string, bool) {
	dests, ok := c.Options["STAGING_JOB_DEST"]
	if !ok {
		return nil, ok
	}
	return strutil.TrimSlice(strutil.Split(dests, ",", true)), true
}

// ProdJobAvailDest prod job 可运行的环境
func (c *WorkspaceOpt) ProdJobAvailDest() ([]string, bool) {
	dests, ok := c.Options["PROD_JOB_DEST"]
	if !ok {
		return nil, ok
	}
	return strutil.TrimSlice(strutil.Split(dests, ",", true)), true
}
