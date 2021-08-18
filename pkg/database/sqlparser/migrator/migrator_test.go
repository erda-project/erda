// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
