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

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	iterationdb "github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/internal/core/legacy/model"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

func Test_metricFieldsEtcdKey(t *testing.T) {
	p := &provider{
		Cfg: &config{
			IterationMetricEtcdPrefixKey: "/devops/metrics/iteration/",
		},
	}
	iterKey := p.metricFieldsEtcdKey(2)
	assert.Equal(t, "/devops/metrics/iteration/2", iterKey)
}

func Test_iterationLabelsFunc(t *testing.T) {
	p := &provider{
		orgSet:       &orgCache{cache.New(cache.NoExpiration, cache.NoExpiration)},
		projectSet:   &projectCache{cache.New(cache.NoExpiration, cache.NoExpiration)},
		iterationSet: &iterationCache{cache.New(cache.NoExpiration, cache.NoExpiration)},
	}
	p.orgSet.Set(1, &orgpb.Org{Name: "org1", ID: 1})
	p.projectSet.Set(1, &model.Project{Name: "project1", OrgID: 1, BaseModel: model.BaseModel{ID: 1}})
	labels := p.iterationLabelsFunc(&IterationInfo{
		Iteration: &iterationdb.Iteration{
			BaseModel: dbengine.BaseModel{ID: 1},
			Title:     "iteration1",
			ProjectID: 1,
		},
	})
	assert.Equal(t, "org1", labels["org_name"])
	assert.Equal(t, "project1", labels["project_name"])
	assert.Equal(t, "iteration1", labels["iteration_title"])
}
func TestIterationIDsLabelsFunc(t *testing.T) {
	i := &iterationCollector{}
	iter := &IterationInfo{
		IterationMetricFields: &IterationMetricFields{
			UUID: "some-uuid",
		},
	}
	labels := i.iterationIDsLabelsFunc(iter)

	expectedLabels := map[string]string{
		"uuid":         "some-uuid",
		"metrics_type": "",
		"ids":          "",
	}
	assert.Equal(t, expectedLabels, labels)
}
