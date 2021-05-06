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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePreserveProjects(t *testing.T) {
	key := "PRESERVEPROJECTS"

	assert.Equal(t, map[string]struct{}{
		"1": {}, "2": {}, "3": {},
	}, ParsePreserveProjects(map[string]string{
		key: "1,2,3",
	}, key))

	assert.Equal(t, map[string]struct{}{}, ParsePreserveProjects(nil, key))
}

func TestBuildDcosConstraints(t *testing.T) {
	assert.Equal(t, [][]string{}, BuildDcosConstraints(false, nil, nil, nil))
	assert.Equal(t, [][]string{}, BuildDcosConstraints(false, map[string]string{
		"MATCH_TAGS":   "",
		"EXCLUDE_TAGS": "locked,platform",
	}, nil, nil))
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bany\b.*`},
	}, BuildDcosConstraints(true, nil, nil, nil))
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bany\b.*`},
	}, BuildDcosConstraints(true, map[string]string{}, nil, nil))
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bany\b.*`},
		{"dice_tags", "UNLIKE", `.*\blocked\b.*`},
		{"dice_tags", "UNLIKE", `.*\bplatform\b.*`},
	}, BuildDcosConstraints(true, map[string]string{
		"MATCH_TAGS":   "",
		"EXCLUDE_TAGS": "locked,platform",
	}, nil, nil))
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bany\b.*`},
		{"dice_tags", "UNLIKE", `.*\block1\b.*`},
	}, BuildDcosConstraints(true, map[string]string{
		"EXCLUDE_TAGS": "lock1",
	}, nil, nil))
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bpack1\b.*`},
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bpack2\b.*`},
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bpack3\b.*`},
		{"dice_tags", "UNLIKE", `.*\block1\b.*`},
		{"dice_tags", "UNLIKE", `.*\block2\b.*`},
		{"dice_tags", "UNLIKE", `.*\block3\b.*`},
	}, BuildDcosConstraints(true, map[string]string{
		"MATCH_TAGS":   "pack1,pack2,pack3",
		"EXCLUDE_TAGS": "lock1,lock2,lock3",
	}, nil, nil))

	// test preserve project
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bproject-32\b.*`},
		{"dice_tags", "LIKE", `.*\bt1\b.*`},
		{"dice_tags", "LIKE", `.*\bt2\b.*`},
		{"dice_tags", "LIKE", `.*\bt3\b.*`},
		{"dice_tags", "UNLIKE", `.*\be1\b.*`},
		{"dice_tags", "UNLIKE", `.*\be2\b.*`},
		{"dice_tags", "UNLIKE", `.*\be3\b.*`},
	}, BuildDcosConstraints(true, map[string]string{
		"MATCH_TAGS":   "t1,t2,t3",
		"EXCLUDE_TAGS": "e1,e2,e3",
		"DICE_PROJECT": "32",
	}, map[string]struct{}{
		"32": {},
	}, nil))

	// test not preserve project
	assert.Equal(t, [][]string{
		{"dice_tags", "UNLIKE", `.*\bproject-[^,]+\b.*`},
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bt1\b.*`},
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bt2\b.*`},
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bt3\b.*`},
		{"dice_tags", "UNLIKE", `.*\be1\b.*`},
		{"dice_tags", "UNLIKE", `.*\be2\b.*`},
		{"dice_tags", "UNLIKE", `.*\be3\b.*`},
	}, BuildDcosConstraints(true, map[string]string{
		"MATCH_TAGS":   "t1,t2,t3",
		"EXCLUDE_TAGS": "e1,e2,e3",
		"DICE_PROJECT": "32",
	}, map[string]struct{}{}, nil))
}

func TestCombineLabels(t *testing.T) {
	assert.Equal(t, map[string]string{}, CombineLabels(nil, nil))

	assert.Equal(t, map[string]string{
		"A": "v1",
		"B": "v4",
		"C": "v3",
	}, CombineLabels(map[string]string{
		"A": "v1",
		"B": "v2",
	}, map[string]string{
		"C": "v3",
		"B": "v4",
	}))
}
