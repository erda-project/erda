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

package dag

import (
	"reflect"
	"testing"
)

func testGraph(t *testing.T) *DAG {

	//  b     a     c
	//  |    / \
	//  |   |   x
	//  |   | / |
	//  |   y   |
	//   \ /    z
	//    w

	t.Helper()
	g, err := New([]NamedNode{
		&MyNode{name: "a"},
		&MyNode{name: "b"},
		&MyNode{name: "c"},
		&MyNode{name: "x", runAfter: []string{"a"}},
		&MyNode{name: "y", runAfter: []string{"a", "x"}},
		&MyNode{name: "z", runAfter: []string{"x"}},
		&MyNode{name: "w", runAfter: []string{"b", "y"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestGetSchedulable(t *testing.T) {
	g := testGraph(t)
	tcs := []struct {
		caseName    string
		finished    []string
		schedulable []string
	}{{
		caseName:    "nothing-done",
		finished:    []string{},
		schedulable: []string{"a", "b", "c"},
	}, {
		caseName:    "a-done",
		finished:    []string{"a"},
		schedulable: []string{"b", "c", "x"},
	}, {
		caseName:    "b-done",
		finished:    []string{"b"},
		schedulable: []string{"a", "c"},
	}, {
		caseName:    "a-and-b-done",
		finished:    []string{"a", "b"},
		schedulable: []string{"c", "x"},
	}, {
		caseName:    "a-x-done",
		finished:    []string{"a", "x"},
		schedulable: []string{"b", "c", "y", "z"},
	}, {
		caseName:    "a-x-b-done",
		finished:    []string{"a", "x", "b"},
		schedulable: []string{"c", "y", "z"},
	}, {
		caseName:    "a-x-y-done",
		finished:    []string{"a", "x", "y"},
		schedulable: []string{"b", "c", "z"},
	}, {
		caseName:    "a-x-y-b-done",
		finished:    []string{"a", "x", "y", "b"},
		schedulable: []string{"c", "w", "z"},
	}}
	for _, tc := range tcs {
		t.Run(tc.caseName, func(t *testing.T) {
			// sort.Sort(sort.StringSlice(tc.schedulable))
			nodeNames, err := g.GetSchedulableNodeNames(tc.finished...)
			if err != nil {
				t.Fatalf("didn't expect error when getting next tasks for %v but got %v", tc.finished, err)
			}
			if len(nodeNames) != len(tc.schedulable) {
				t.Fail()
			}
			if !reflect.DeepEqual(nodeNames, tc.schedulable) {
				t.Fatalf("schedulable: expected %v, actual %v", tc.schedulable, nodeNames)
			}
		})
	}
}

func TestGetSchedulable_Invalid(t *testing.T) {
	g := testGraph(t)
	tcs := []struct {
		name     string
		finished []string
	}{{
		name:     "only-x",
		finished: []string{"x"},
	}, {
		name:     "only-y",
		finished: []string{"y"},
	}, {
		name:     "only-w",
		finished: []string{"w"},
	}, {
		name:     "only-y-and-x",
		finished: []string{"y", "x"},
	}, {
		name:     "only-y-and-w",
		finished: []string{"y", "w"},
	}, {
		name:     "only-x-and-w",
		finished: []string{"x", "w"},
	}}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := g.GetSchedulable(tc.finished...)
			if err == nil {
				t.Fatalf("expected error for invalid done tasks %v but got none", tc.finished)
			}
		})
	}
}
