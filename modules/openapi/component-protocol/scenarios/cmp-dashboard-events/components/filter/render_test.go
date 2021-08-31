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

package filter

import "testing"

func TestHasSuffix(t *testing.T) {
	if _, ok := hasSuffix("project-1-dev"); !ok {
		t.Errorf("test failed, \"project-1-dev\" has project suffix, actual not")
	}
	if _, ok := hasSuffix("project-2-staging"); !ok {
		t.Errorf("test failed, \"project-2-staging\" has project suffix, actual not")
	}
	if _, ok := hasSuffix("project-3-test"); !ok {
		t.Errorf("test failed, \"project-3-test\" has project suffix, actual not")
	}
	if _, ok := hasSuffix("project-4-prod"); !ok {
		t.Errorf("test failed, \"project-4-prod\" has project suffix, actual not")
	}
	if _, ok := hasSuffix("project-5-custom"); ok {
		t.Errorf("test failed, \"project-5-custom\" does not have project suffix, actul do")
	}
}
