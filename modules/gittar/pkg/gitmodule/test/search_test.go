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

package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Branch represents a Git branch.
type Branch struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
	IsProtect bool   `json:"isProtect"`
	IsMerged  bool   `json:"isMerged"` // 是否已合并到默认分支
}

func TestBranch(t *testing.T) {
	branches := []*Branch{
		{
			Id:        "c496b91dc7590ebe8eb810be87e8704669302a8c",
			Name:      "master",
			IsDefault: true,
			IsProtect: false,
			IsMerged:  true,
		},
		{
			Id:        "18d2c31b49e4e0d725bc9709e1586e6f744bdc0c",
			Name:      "tt",
			IsDefault: false,
			IsProtect: false,
			IsMerged:  false,
		},
	}
	newBranches := []*Branch{}
	for _, branch := range branches {
		if !strings.Contains(branch.Name, "t") {
			continue
		}
		newBranches = append(newBranches, branch)
	}
	assert.Equal(t, newBranches, []*Branch{
		{
			Id:        "c496b91dc7590ebe8eb810be87e8704669302a8c",
			Name:      "master",
			IsDefault: true,
			IsProtect: false,
			IsMerged:  true,
		},
		{
			Id:        "18d2c31b49e4e0d725bc9709e1586e6f744bdc0c",
			Name:      "tt",
			IsDefault: false,
			IsProtect: false,
			IsMerged:  false,
		},
	})
}

func TestOnlyBranchNames(t *testing.T) {
	branches := []string{"master", "tt", "aaa"}
	newBranches := []*Branch{}
	for _, branchName := range branches {
		if !strings.Contains(branchName, "t") {
			continue
		}
		newBranches = append(newBranches, &Branch{Name: branchName})
	}
	assert.Equal(t, newBranches, []*Branch{
		{
			Id:        "",
			Name:      "master",
			IsDefault: false,
			IsProtect: false,
			IsMerged:  false,
		},
		{
			Id:        "",
			Name:      "tt",
			IsDefault: false,
			IsProtect: false,
			IsMerged:  false,
		},
	})
}
