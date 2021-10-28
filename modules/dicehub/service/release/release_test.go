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

package release

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/erda-project/erda/apistructs"
)

func TestLimitLabelsLength(t *testing.T) {
	req1 := &apistructs.ReleaseCreateRequest{
		Labels: nil,
	}
	if err := limitLabelsLength(req1); err != nil {
		t.Error(err)
	}

	req2 := &apistructs.ReleaseCreateRequest{
		Labels: map[string]string{
			"a": rand.String(100),
			"b": rand.String(101),
			"c": rand.String(98) + "中文的",
		},
	}
	if err := limitLabelsLength(req2); err != nil {
		t.Error(err)
	}

	req3 := &apistructs.ReleaseCreateRequest{
		Labels: map[string]string{
			"a": rand.String(1000),
			"b": rand.String(100),
			"c": rand.String(98) + "中文的",
		},
	}
	if err := limitLabelsLength(req3); err != nil {
		t.Error(err)
	}
	for _, v := range req3.Labels {
		if len([]rune(v)) > 100 {
			t.Error("fail")
		}
	}
}
