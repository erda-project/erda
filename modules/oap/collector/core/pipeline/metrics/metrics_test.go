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

package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/stretchr/testify/assert"
)

func TestPipeline_StartStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pipe := NewPipeline(logrusx.New().Sub("collector"))

	// invalid
	assert.Error(t, pipe.InitComponents([]model.Component{&model.NoopMetricProcessor{}}, nil, nil))

	// normal
	err := pipe.InitComponents([]model.Component{&model.NoopMetricReceiver{}}, []model.Component{&model.NoopMetricProcessor{}}, []model.Component{&model.NoopMetricExporter{}})
	assert.Nil(t, err)

	pipe.StartStream(ctx)

	time.Sleep(time.Second)
}
