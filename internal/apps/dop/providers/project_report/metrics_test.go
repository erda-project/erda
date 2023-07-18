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

package project_report

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
)

func Test_generateProjectMetricLabels(t *testing.T) {
	projectDto := &apistructs.ProjectDTO{
		ID:          1,
		Name:        "project1",
		OrgID:       1,
		DisplayName: "project1",
	}
	orgDto := &pb.Org{
		ID:          1,
		Name:        "org1",
		DisplayName: "org1",
	}
	keys, values, labels := generateProjectMetricLabels(projectDto, orgDto)
	assert.Equal(t, 8, len(keys))
	assert.Equal(t, 8, len(values))
	assert.Equal(t, "org1", labels[labelOrgName])
}
