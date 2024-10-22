// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package efficiency_measure

import "context"

func (p *provider) refreshBasicInfo() error {
	if len(p.Cfg.OrgWhiteList) == 0 {
		return nil
	}
	
	userProjects, err := p.projDB.GetAllUserJoinedProjects(p.Cfg.OrgWhiteList)
	if err != nil {
		p.Log.Errorf("failed to get all user joined projects, err: %v", err)
		return err
	}
	for i := range userProjects {
		if personalInfo := p.personalEfficiencySet.Get(userProjects[i].ID); personalInfo != nil {
			personalInfo.userProject = userProjects[i]
		} else {
			p.personalEfficiencySet.Set(userProjects[i].ID, &PersonalPerformanceInfo{
				userProject: userProjects[i],
			})
		}
	}
	return nil
}

func (p *provider) GetRequestedPersonalInfos() (map[uint64]*PersonalPerformanceInfo, error) {
	result := make(map[uint64]*PersonalPerformanceInfo)
	p.personalEfficiencySet.Iterate(func(key string, value interface{}) error {
		personalInfo := value.(*PersonalPerformanceInfo)
		if personalInfo.metricFields == nil {
			fields := &personalMetricField{}
			if err := p.Js.Get(context.Background(), p.metricFieldsEtcdKey(personalInfo.userProject.ID), fields); err == nil && fields.IsValid() {
				personalInfo.metricFields = fields
			}
		}
		result[personalInfo.userProject.ID] = personalInfo
		return nil
	})
	return result, nil
}
