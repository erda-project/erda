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

package instanceinfosync

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_extractEnvs(t *testing.T) {
	pod := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{},
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"addon.erda.cloud/id":          "10001",
				"core.erda.cloud/cluster-name": "erda-jicheng",
				"core.erda.cloud/org-name":     "development-erda",
				"core.erda.cloud/org-id":       "888",
				"core.erda.cloud/project-name": "Master",
				"core.erda.cloud/project-id":   "8888",
				"core.erda.cloud/app-name":     "go-demo",
				"core.erda.cloud/app-id":       "88888",
				"core.erda.cloud/runtime-id":   "66666",
				"core.erda.cloud/service-name": "go-demo-web",
				"core.erda.cloud/workspace":    "prod",
				"addon.erda.cloud/type":        "mysql",
			},
		},
	}

	extractEnvs(pod)
}
