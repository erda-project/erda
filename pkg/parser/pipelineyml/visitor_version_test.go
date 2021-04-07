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

package pipelineyml

import (
	"testing"
)

func TestVersionVisitor_Visit(t *testing.T) {
	v := NewVersionVisitor()

	validVersionTestCases := []string{
		`version: 1.1`,
		`version: '1.1'`,
		`version: "1.1"`,
	}
	for _, tc := range validVersionTestCases {
		y, err := New([]byte(tc))
		if err != nil {
			t.Fatal(err)
		}
		y.s.Accept(v)
		if len(y.s.errs) > 0 {
			t.Fatal(y.s.mergeErrors())
		}
	}

	invalidVersionTestCases := []string{
		`version: 1.2`,
		`version: 2`,
		`version: 1.1.alpha`,
		`version: "1`,
		`version: test`,
	}
	for _, tc := range invalidVersionTestCases {
		_, err := New([]byte(tc))
		if err == nil {
			t.Fatalf("should error: `%s`", tc)
		}
	}
}

func TestGetVersion(t *testing.T) {
	validVersionTestCases := []string{
		`version: 1`,
		`version: '1'`,
		`version: "1"`,
		`version: 1.0`,
		`version: '1.0'`,
		`version: "1.0"`,
	}
	for _, tc := range validVersionTestCases {
		version, err := GetVersion([]byte(tc))
		if err != nil {
			t.Fatal(err)
		}
		if !(version == "1" || version == "1.0") {
			t.Fatalf("invalid version: %s", version)
		}
	}
}
