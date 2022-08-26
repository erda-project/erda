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

package autotestv2

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/apps/dop/dao"
)

func Test_replaceNewStepValue(t *testing.T) {
	type arg struct {
		stepIDMap map[uint64]uint64
		stepValue string
	}
	testCases := []struct {
		name     string
		arg      arg
		expected string
	}{
		{
			name: "no loop",
			arg: arg{
				stepValue: `{
    "apiSpec": {
        "url": "https://erda.cloud"
    },
}`,
				stepIDMap: map[uint64]uint64{
					52203: 52204,
				},
			},
			expected: `{
    "apiSpec": {
        "url": "https://erda.cloud"
    },
}`,
		},
		{
			name: "loop with self id",
			arg: arg{
				stepValue: `{
    "apiSpec": {
        "url": "https://erda.cloud"
    },
    "loop": {
        "break": "'${{ outputs.52203.status }}' == '200'",
        "strategy": {
            "decline_limit_sec": 3,
            "decline_ratio": 2,
            "interval_sec": 1,
            "max_times": 5
        }
    }
}`,
				stepIDMap: map[uint64]uint64{
					52203: 52204,
				},
			},
			expected: `{
    "apiSpec": {
        "url": "https://erda.cloud"
    },
    "loop": {
        "break": "'${{ outputs.52204.status }}' == '200'",
        "strategy": {
            "decline_limit_sec": 3,
            "decline_ratio": 2,
            "interval_sec": 1,
            "max_times": 5
        }
    }
}`,
		},
	}
	db := &dao.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "UpdateAutotestSceneStep", func(_ *dao.DBClient, step *dao.AutoTestSceneStep) error {
		return nil
	})
	defer pm1.Unpatch()
	svc := &Service{db: db}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newStep := &dao.AutoTestSceneStep{
				Value: tc.arg.stepValue,
			}
			err := svc.replaceNewStepValue(newStep, tc.arg.stepIDMap)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, newStep.Value)
		})
	}
}
