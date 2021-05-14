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

package project

import "testing"

func TestClass_genProjectNamespace(t *testing.T) {
	namespaces := genProjectNamespace("1")
	expected := map[string]string{"DEV": "project-1-dev", "TEST": "project-1-test", "STAGING": "project-1-staging", "PROD": "project-1-prod"}
	for env, expectedNS := range expected {
		actuallyNS, ok := namespaces[env]
		if !ok {
			t.Errorf("env not existd: %s", env)
		}
		if expectedNS != actuallyNS {
			t.Errorf("expected: [%s:%s], actually: [%s:%s]", env, expectedNS, env, actuallyNS)
		}
	}
}
