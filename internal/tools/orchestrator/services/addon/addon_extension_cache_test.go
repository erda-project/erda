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

package addon

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func Test_GetDefault(t *testing.T) {
	defer monkey.UnpatchAll()

	InitCache(DefaultTtl, DefaultSize)

	monkey.PatchInstanceMethod(reflect.TypeOf(cache.bdl), "QueryExtensionVersions", func(_ *bundle.Bundle,
		req apistructs.ExtensionVersionQueryRequest) ([]apistructs.ExtensionVersion, error) {
		switch req.Name {
		case "mysql":
			return []apistructs.ExtensionVersion{
				{Name: "mysql", Version: "8.0.0", IsDefault: true},
				{Name: "mysql", Version: "5.7.2"},
			}, nil
		case "canal":
			return []apistructs.ExtensionVersion{
				{Name: "canal", Version: "1.1.5"},
				{Name: "canal", Version: "1.1.6"},
			}, nil
		case "custom":
			return []apistructs.ExtensionVersion{
				{Name: "custom", Version: "1.0.0"},
			}, nil
		default:
			return []apistructs.ExtensionVersion{}, errors.New("not found")
		}
	})

	type args struct {
		name string
	}

	tests := []struct {
		name          string
		args          args
		expectVersion string
		expectEmpty   bool
	}{
		{
			name: "multi version, get default version",
			args: args{
				name: "mysql",
			},
			expectVersion: "8.0.0",
		},
		{
			name: "one version, get default version",
			args: args{
				name: "custom",
			},
			expectVersion: "1.0.0",
		},
		{
			name: "get non default version",
			args: args{
				name: "canal",
			},
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheMap, err := GetCache().Get(tt.args.name)
			assert.NoError(t, err, "Expected no error, but got ", err)

			versionMap := cacheMap.(*VersionMap)
			defaultVersion, ok := versionMap.GetDefault()
			assert.Equal(t, !ok, tt.expectEmpty,
				fmt.Sprintf("Expected %v, but got %v", tt.expectEmpty, ok))
			assert.Equal(t, tt.expectVersion, defaultVersion.Version,
				fmt.Sprintf("Expected %s, but got %s", tt.expectVersion, defaultVersion.Version))
		})
	}
}
