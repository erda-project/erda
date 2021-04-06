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

package sysconf

import (
	"testing"
)

func TestIsClusterName(t *testing.T) {
	m := map[string]bool{
		"":              false,
		"terminus":      false,
		"terminus-test": true,
		"terminus-prod": true,
	}
	for k, v := range m {
		if v != isClusterName(k) {
			t.Fatal(k)
		}
	}
}

func TestIsPort(t *testing.T) {
	m := map[int]bool{
		0:     false,
		-1:    false,
		22:    true,
		65535: true,
		65536: false,
	}
	for k, v := range m {
		if v != IsPort(k) {
			t.Fatal(k)
		}
	}
}

func TestIsDNSName(t *testing.T) {
	m := map[string]bool{
		"":            false,
		"*.aa":        false,
		"a.b":         true,
		"192.168.0.1": true,
		"test.com":    true,
	}
	for k, v := range m {
		if v != IsDNSName(k) {
			t.Fatal(k)
		}
	}
}
