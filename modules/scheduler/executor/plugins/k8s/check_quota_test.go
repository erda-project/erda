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

package k8s

import "testing"

func TestResourceToString(t *testing.T) {
	cpu := 1000.0
	mem := float64(1 << 30)
	cpuStr := resourceToString(cpu, "cpu")
	memStr := resourceToString(mem, "memory")
	if cpuStr != "1" {
		t.Errorf("test failed, expected cpu is \"1\", got %s", cpuStr)
	}
	if memStr != "1G" {
		t.Errorf("test failed, expected cpu is \"1G\", got %s", memStr)
	}
}
