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

package executortypes

import (
	"strconv"

	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/pkg/strutil"
)

// EnableLabelSchedule Whether tag scheduling is turned on
func (c *ExecutorWholeConfigs) EnableLabelSchedule() bool {
	if v, ok := c.BasicConfig["ENABLETAG"]; ok && len(v) > 0 {
		if enable, err := strconv.ParseBool(v); err == nil {
			return enable
		}
	}
	return false
}

// ProjectIDForCompatibility Isolate the configuration used by the project when scheduling, and no longer use it
// DEPRECATED
func (c *ExecutorWholeConfigs) ProjectIDForCompatibility(projectID string) bool {
	_, exist := c.BasicConfig[projectID]
	return exist
}

// EnableOrgLabelSchedule Whether to enable org-level label scheduling
func (c *ExecutorWholeConfigs) EnableOrgLabelSchedule() bool {
	if !c.EnableLabelSchedule() {
		return false
	}
	if enable, ok := c.BasicConfig["ENABLE_ORG"]; ok && strutil.ToLower(enable) == "true" {
		return true
	}
	return false
}

// EnableWorkspaceLabelSchedule Whether to enable label scheduling at the workspace level
func (c *ExecutorWholeConfigs) EnableWorkspaceLabelSchedule() bool {
	if !c.EnableLabelSchedule() {
		return false
	}
	if enable, ok := c.BasicConfig["ENABLE_WORKSPACE"]; ok && strutil.ToLower(enable) == "true" {
		return true
	}
	return false
}

// WORKSPACETAGSForCompatibility Compatible with old WORKSPACETAGS tags
// DEPRECATED
func (c *ExecutorWholeConfigs) WORKSPACETAGSForCompatibility() (string, bool) {
	v, ok := c.BasicConfig["WORKSPACETAGS"]
	return v, ok
}

// StagingJobAvailDest staging environment where job can run
func (c *ExecutorWholeConfigs) StagingJobAvailDest() ([]string, bool) {
	dests, ok := c.BasicConfig["STAGING_JOB_DEST"]
	if !ok {
		return nil, ok
	}
	return strutil.TrimSlice(strutil.Split(dests, ",", true)), true
}

// ProdJobAvailDest prod environment where job can run
func (c *ExecutorWholeConfigs) ProdJobAvailDest() ([]string, bool) {
	dests, ok := c.BasicConfig["PROD_JOB_DEST"]
	if !ok {
		return nil, ok
	}
	return strutil.TrimSlice(strutil.Split(dests, ",", true)), true
}

// OrgOpt org level configuration
type OrgOpt conf.Org

// WorkspaceOpt workspace level configuration
type WorkspaceOpt conf.Workspace

// OrgOpt Take out the configuration of `org' from all configurations
func (c *ExecutorWholeConfigs) OrgOpt(org string) *OrgOpt {
	for _, orgopt := range c.PlusConfigs.Orgs {
		if orgopt.Name == org {
			opt := OrgOpt(orgopt)
			return &opt
		}
	}
	return nil
}

// WorkspaceOpt Take out the configuration of `workspace' from the org configuration
func (c *OrgOpt) WorkspaceOpt(workspace string) *WorkspaceOpt {
	for _, workspaceopt := range c.Workspaces {
		if workspaceopt.Name == workspace {
			opt := WorkspaceOpt(workspaceopt)
			return &opt
		}
	}
	return nil
}

// EnableWorkspaceLabelSchedule Whether to enable workspace label scheduling
func (c *WorkspaceOpt) EnableWorkspaceLabelSchedule() bool {
	if enable, ok := c.Options["ENABLE_WORKSPACE"]; ok && strutil.ToLower(enable) == "true" {
		return true
	}
	return false
}

// StagingJobAvailDest Environment where staging job can run
func (c *WorkspaceOpt) StagingJobAvailDest() ([]string, bool) {
	dests, ok := c.Options["STAGING_JOB_DEST"]
	if !ok {
		return nil, ok
	}
	return strutil.TrimSlice(strutil.Split(dests, ",", true)), true
}

// ProdJobAvailDest Environment where prod job can run
func (c *WorkspaceOpt) ProdJobAvailDest() ([]string, bool) {
	dests, ok := c.Options["PROD_JOB_DEST"]
	if !ok {
		return nil, ok
	}
	return strutil.TrimSlice(strutil.Split(dests, ",", true)), true
}
