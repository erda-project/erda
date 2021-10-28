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
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestCopy(t *testing.T) {
	a := &AutoTestSpaceData{
		Space:        &apistructs.AutoTestSpace{Status: "open"},
		IsCopy:       true,
		NewSpace:     &apistructs.AutoTestSpace{},
		IdentityInfo: apistructs.IdentityInfo{UserID: "1"},
		Steps: map[uint64][]apistructs.AutoTestSceneStep{
			1: []apistructs.AutoTestSceneStep{},
		},
		Scenes: map[uint64][]apistructs.AutoTestScene{
			1: []apistructs.AutoTestScene{},
		},
	}
	m1 := monkey.PatchInstanceMethod(reflect.TypeOf(a), "CreateNewSpace",
		func(a *AutoTestSpaceData) error {
			return nil
		})
	defer m1.Unpatch()

	autotestSvc := New()
	a.svc = autotestSvc

	m2 := monkey.PatchInstanceMethod(reflect.TypeOf(autotestSvc), "UpdateAutoTestSpace",
		func(svc *Service, req apistructs.AutoTestSpace, UserID string) (*apistructs.AutoTestSpace, error) {
			return &apistructs.AutoTestSpace{}, nil
		})
	defer m2.Unpatch()

	m3 := monkey.PatchInstanceMethod(reflect.TypeOf(a), "CopySceneSets",
		func(a *AutoTestSpaceData) error {
			return nil
		})
	defer m3.Unpatch()

	m4 := monkey.PatchInstanceMethod(reflect.TypeOf(a), "CopyScenes",
		func(a *AutoTestSpaceData) error {
			return nil
		})
	defer m4.Unpatch()

	m5 := monkey.PatchInstanceMethod(reflect.TypeOf(a), "CopySceneSteps",
		func(a *AutoTestSpaceData) error {
			return nil
		})
	defer m5.Unpatch()

	m6 := monkey.PatchInstanceMethod(reflect.TypeOf(a), "CopyInputs",
		func(a *AutoTestSpaceData) error {
			if len(a.Steps) == 0 {
				return fmt.Errorf("invalid steps")
			}
			return nil
		})
	defer m6.Unpatch()

	m7 := monkey.PatchInstanceMethod(reflect.TypeOf(a), "CopyOutputs",
		func(a *AutoTestSpaceData) error {
			if len(a.Scenes) == 0 {
				return fmt.Errorf("invalid steps")
			}
			return nil
		})
	defer m7.Unpatch()

	_, err := a.Copy()
	assert.NoError(t, err)
	emptyStepsData := &AutoTestSpaceData{
		Space:        &apistructs.AutoTestSpace{Status: "open"},
		IsCopy:       true,
		NewSpace:     &apistructs.AutoTestSpace{},
		IdentityInfo: apistructs.IdentityInfo{UserID: "1"},
	}
	_, err = emptyStepsData.Copy()
	assert.Equal(t, false, err == nil)
	emptyScenesData := &AutoTestSpaceData{
		Space:        &apistructs.AutoTestSpace{Status: "open"},
		IsCopy:       true,
		NewSpace:     &apistructs.AutoTestSpace{},
		IdentityInfo: apistructs.IdentityInfo{UserID: "1"},
		Steps: map[uint64][]apistructs.AutoTestSceneStep{
			1: []apistructs.AutoTestSceneStep{},
		},
	}
	_, err = emptyScenesData.Copy()
	assert.Equal(t, false, err == nil)
}
