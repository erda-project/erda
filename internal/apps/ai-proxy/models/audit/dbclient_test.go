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

package audit

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
)

func TestJsonRawMessage(t *testing.T) {
	raw := json.RawMessage(time.Now().Sub(time.Now().Add(-time.Second)).String())
	fmt.Println(string(raw))
}

func TestValidateAndGetTimeRange(t *testing.T) {
	now := time.Now()

	t.Run("default to last 24 hours", func(t *testing.T) {
		req := &pb.AuditPagingRequest{}
		before, after, err := ValidateAndGetTimeRange(req)
		assert.NoError(t, err)
		assert.WithinDuration(t, now, before, time.Second)
		assert.WithinDuration(t, now.AddDate(0, 0, -1), after, time.Second)
	})

	t.Run("both time params provided within 24 hours", func(t *testing.T) {
		beforeMs := now.UnixNano() / 1e6
		afterMs := now.Add(-12 * time.Hour).UnixNano() / 1e6
		req := &pb.AuditPagingRequest{
			TimeRangeBeforeMs: uint64(beforeMs),
			TimeRangeAfterMs:  uint64(afterMs),
		}
		before, after, err := ValidateAndGetTimeRange(req)
		assert.NoError(t, err)
		assert.Equal(t, time.Unix(beforeMs/1000, (beforeMs%1000)*1e6).UTC(), before.UTC())
		assert.Equal(t, time.Unix(afterMs/1000, (afterMs%1000)*1e6).UTC(), after.UTC())
	})

	t.Run("time range exceeds 24 hours", func(t *testing.T) {
		beforeMs := now.UnixNano() / 1e6
		afterMs := now.Add(-25 * time.Hour).UnixNano() / 1e6
		req := &pb.AuditPagingRequest{
			TimeRangeBeforeMs: uint64(beforeMs),
			TimeRangeAfterMs:  uint64(afterMs),
		}
		_, _, err := ValidateAndGetTimeRange(req)
		assert.Error(t, err)
		assert.Equal(t, "time range span cannot exceed one day", err.Error())
	})

	t.Run("only one time param provided", func(t *testing.T) {
		req := &pb.AuditPagingRequest{
			TimeRangeBeforeMs: uint64(now.UnixNano() / 1e6),
		}
		_, _, err := ValidateAndGetTimeRange(req)
		assert.Error(t, err)
		assert.Equal(t, "TimeRangeBeforeMs and TimeRangeAfterMs must be passed together", err.Error())
	})

	t.Run("before is after", func(t *testing.T) {
		beforeMs := now.Add(-25 * time.Hour).UnixNano() / 1e6
		afterMs := now.UnixNano() / 1e6
		req := &pb.AuditPagingRequest{
			TimeRangeBeforeMs: uint64(beforeMs),
			TimeRangeAfterMs:  uint64(afterMs),
		}
		_, _, err := ValidateAndGetTimeRange(req)
		assert.Error(t, err)
		assert.Equal(t, "TimeRangeBeforeMs must be after TimeRangeAfterMs", err.Error())
	})
}