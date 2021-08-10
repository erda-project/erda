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
