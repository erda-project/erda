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

package migrator

import "testing"

func TestMigrator_needCompare(t *testing.T) {
	mig := new(Migrator)
	mig.installingType = firstTimeInstall
	if mig.needCompare() {
		t.Fatal(firstTimeInstall, "need not to compare")
	}

	mig.installingType = normalUpdate
	if mig.needCompare() {
		t.Fatal(normalUpdate, "need not to compare")
	}

	mig.installingType = firstTimeUpdate
	if !mig.needCompare() {
		t.Fatal(firstTimeUpdate, "need to compare")
	}
}
