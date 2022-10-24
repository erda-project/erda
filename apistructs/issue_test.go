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

package apistructs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetRelatedIssueIDs(t *testing.T) {
	iss := Issue{}
	iss.SetRelatedIssueIDs("1001,1002")
	relatedIDs := iss.GetRelatedIssueIDs()
	assert.Equal(t, 2, len(relatedIDs))
	assert.Equal(t, uint64(1001), relatedIDs[0])
	assert.Equal(t, uint64(1002), relatedIDs[1])
}

func TestNewManhour(t *testing.T) {
	_, err := NewManhour("2w3d")
	assert.Equal(t, true, err != nil)
	est1, _ := NewManhour("3h")
	assert.Equal(t, int64(180), est1.EstimateTime)
	est2, _ := NewManhour("9m")
	assert.Equal(t, int64(9), est2.EstimateTime)
	est3, _ := NewManhour("5d")
	assert.Equal(t, int64(2400), est3.EstimateTime)
	est4, _ := NewManhour("1w")
	assert.Equal(t, int64(2400), est4.EstimateTime)
}

func TestIssueTime_IsEmpty(t *testing.T) {
	t1 := IssueTime(time.Unix(0, 0))
	tests := []struct {
		name string
		m    *IssueTime
		want bool
	}{
		{
			m:    &t1,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &IssueTime{}
			if got := m.IsEmpty(); got != tt.want {
				t.Errorf("IssueTime.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
