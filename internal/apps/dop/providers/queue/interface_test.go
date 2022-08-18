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

package queue

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func Test_makeMustMatchLabels(t *testing.T) {
	labels := makeMustMatchLabels("DEV", &apistructs.ProjectDTO{
		ID:   1,
		Name: "queue",
	})
	assert.Equal(t, 3, len(labels))
	assert.Equal(t, fmt.Sprintf("%s=DEV", projectQueueLabelKeyWorkspace), labels[1])
}

func Test_makeProjectLevelQueueName(t *testing.T) {
	p := &provider{}
	name := p.makeProjectLevelQueueName("DEV", "queue")
	assert.Equal(t, "project-queue-DEV", name)
}
