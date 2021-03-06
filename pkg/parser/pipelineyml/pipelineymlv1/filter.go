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

package pipelineymlv1

import (
	"regexp"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

type Filters []Filter

type Filter struct {
	Type    string   `json:"type"`
	Key     string   `json:"key"` // used for "global-env" filter type
	Onlys   []string `json:"onlys,omitempty" mapstructure:"onlys"`
	Excepts []string `json:"excepts,omitempty" mapstructure:"excepts"`
}

const (
	GIT_BRANCH = "git-branch"
	GLOBAL_ENV = "global-env"
)

func (filter Filter) parse() error {
	var me error
	for _, only := range filter.Onlys {
		if _, err := regexp.Compile(only); err != nil {
			me = multierror.Append(me, err)
		}
	}
	for _, except := range filter.Excepts {
		if _, err := regexp.Compile(except); err != nil {
			me = multierror.Append(me, err)
		}
	}
	return me
}

func (filters Filters) parse() error {
	var me error
	for _, f := range filters {
		if err := f.parse(); err != nil {
			me = multierror.Append(me, err)
		}
	}
	return me
}

func (filters Filters) needDisable(gitBranch string, globalEnvs map[string]string) bool {
	if filters == nil {
		return false
	}
	for _, filter := range filters {
		if filter.needDisable(gitBranch, globalEnvs) {
			return true
		}
	}
	return false
}

// needDisable return disable or not
// 默认都是 enabled 的
func (filter *Filter) needDisable(gitBranch string, globalEnvs map[string]string) bool {
	switch filter.Type {
	case GIT_BRANCH:
		return filter.doDisable(gitBranch)
	case GLOBAL_ENV:
		v, ok := globalEnvs[filter.Key]
		if !ok {
			return false
		}
		return filter.doDisable(v)
	default:
		return false
	}
}

// Any branches that match only will run the job.
// Any branches that match ignore will not run the job.
// If neither only nor ignore are specified then all branches will run the job.
// If both only and ignore are specified the only is considered before ignore.
func (filter *Filter) doDisable(input string) (needDisable bool) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("failed to compile filter regexp of onlys or excepts. panic: %v", r)
			needDisable = false
		}
	}()

	// match onlys
	var needDisableByOnly = true
	if len(filter.Onlys) == 0 {
		needDisableByOnly = false
	} else {
		for _, only := range filter.Onlys {
			if regexp.MustCompile(only).MatchString(input) {
				needDisableByOnly = false
				break
			}
		}
	}

	// match excepts
	var needDisableByExcept = false
	for _, except := range filter.Excepts {
		if regexp.MustCompile(except).MatchString(input) {
			needDisableByExcept = true
			break
		}
	}

	return needDisableByOnly || needDisableByExcept
}
