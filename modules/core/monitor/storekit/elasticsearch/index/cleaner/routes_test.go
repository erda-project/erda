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

package cleaner

import (
	"context"
	"net/http"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
)

func Test_cleanExpiredIndices_With_EmptyTimeOffset_Should_UseDefault(t *testing.T) {
	p := &provider{}
	nowTime := time.Now()
	var actualTime time.Time

	monkey.Patch(time.Now, func() time.Time {
		return nowTime
	})
	defer monkey.Unpatch(time.Now)
	monkey.Patch((*provider).checkAndDeleteIndices, func(p *provider, ctx context.Context, now time.Time, filter loader.Matcher) error {
		actualTime = now
		return nil
	})
	defer monkey.Unpatch((*provider).checkAndDeleteIndices)

	r, _ := http.NewRequest("GET", "url", nil)
	result := p.cleanExpiredIndices(r, struct {
		TimeOffset string `query:"timeOffset"`
	}{
		TimeOffset: "",
	})

	assert.Equal(t, true, result)
	assert.Equal(t, nowTime, actualTime)
}

func Test_cleanExpiredIndices_With_ValidTimeOffset_Should_AddTime(t *testing.T) {
	p := &provider{}

	nowTime := time.Now()
	var actualTime time.Time

	monkey.Patch(time.Now, func() time.Time {
		return nowTime
	})
	defer monkey.Unpatch(time.Now)

	monkey.Patch((*provider).checkAndDeleteIndices, func(p *provider, ctx context.Context, now time.Time, filter loader.Matcher) error {
		actualTime = now
		return nil
	})
	defer monkey.Unpatch((*provider).checkAndDeleteIndices)

	r, _ := http.NewRequest("GET", "url", nil)
	result := p.cleanExpiredIndices(r, struct {
		TimeOffset string `query:"timeOffset"`
	}{
		TimeOffset: "1h",
	})

	assert.Equal(t, true, result)
	assert.Equal(t, nowTime.Add(time.Hour), actualTime)
}
