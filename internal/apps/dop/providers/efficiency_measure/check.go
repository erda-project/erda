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

func (p *provider) checkPersonalNumberFields() {
	p.personalEfficiencySet.Iterate(func(key string, value interface{}) error {
		personalInfo := value.(*PersonalPerformanceInfo)
		if personalInfo.metricFields == nil || !personalInfo.metricFields.IsValid() {
			fields, err := p.getPersonalMetricFields(personalInfo)
			if err != nil {
				p.Log.Errorf("failed to generate personal metric fields, projectID: %d, userID: %d, err: %v",
					personalInfo.userProject.ProjectID, personalInfo.userProject.UserID, err)
				return nil
			}
			personalInfo.metricFields = fields
		}
		return nil
	})
}
