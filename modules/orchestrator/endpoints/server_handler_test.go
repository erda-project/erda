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

package endpoints

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func Test_GenOverlayDataForAudit(t *testing.T) {
	oldServiceData := &diceyml.Service{
		Resources: diceyml.Resources{
			CPU:  1,
			Mem:  1024,
			Disk: 0,
		},
		Deployments: diceyml.Deployments{
			Replicas: 1,
		},
	}

	auditData := genOverlayDataForAudit(oldServiceData)

	assert.Equal(t, float64(1), auditData.Resources.CPU)
	assert.Equal(t, 1024, auditData.Resources.Mem)
	assert.Equal(t, 0, auditData.Resources.Disk)
	assert.Equal(t, 1, auditData.Deployments.Replicas)
}
