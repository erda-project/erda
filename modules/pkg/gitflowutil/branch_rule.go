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

package gitflowutil

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

// IsValidBranchWorkspace return if expectWorkspace is valid workspace & artifactWorkspace
func IsValidBranchWorkspace(rules []apistructs.ValidBranch, expectWorkspace apistructs.DiceWorkspace) (validWorkspace, validArtifactWorkspace bool) {
	for _, rule := range rules {
		for _, ws := range strutil.Split(rule.Workspace, ",", true) {
			if ws == expectWorkspace.String() && !validWorkspace {
				validWorkspace = true
			}
		}
		for _, ws := range strutil.Split(rule.ArtifactWorkspace, ",", true) {
			if ws == expectWorkspace.String() && !validArtifactWorkspace {
				validArtifactWorkspace = true
			}
		}
	}
	return
}
