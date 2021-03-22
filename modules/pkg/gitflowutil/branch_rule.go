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
