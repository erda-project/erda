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

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var ISSUEBIND = command.Command{
	ParentName: "ISSUE",
	Name:       "bind",
	ShortHelp:  "bind child issues to a parent issue",
	Example:    "$ erda-cli issue bind --parent 18 --children 101,102",
	Flags: []command.Flag{
		command.StringFlag{Name: "org", Doc: "organization name; defaults to current workspace context", DefaultValue: ""},
		command.StringFlag{Name: "project", Doc: "project name; defaults to current workspace context", DefaultValue: ""},
		command.IntFlag{Name: "parent", Doc: "parent issue id, usually a requirement", DefaultValue: 0},
		command.StringFlag{Name: "children", Doc: "child issue ids, comma-separated", DefaultValue: ""},
		command.StringFlag{Name: "type", Doc: "relation type: inclusion or connection", DefaultValue: apistructs.IssueRelationInclusion},
	},
	Run: IssueBind,
}

var (
	getIssueBindOrgID               = common.GetOrgID
	getIssueBindProjectID           = common.GetProjectID
	bindIssueRelation               = common.CreateIssueRelation
	issueBindStdout       io.Writer = os.Stdout
)

type issueBindOutput struct {
	ParentID uint64   `json:"parentID"`
	Children []uint64 `json:"children"`
	Type     string   `json:"type"`
}

func IssueBind(ctx *command.Context, org string, project string, parent int, children string, relationType string) error {
	if parent <= 0 {
		return fmt.Errorf("--parent must be greater than 0")
	}
	childIDs, err := parseIssueBindChildren(children)
	if err != nil {
		return err
	}
	if relationType == "" {
		relationType = apistructs.IssueRelationInclusion
	}
	if relationType != apistructs.IssueRelationInclusion && relationType != apistructs.IssueRelationConnection {
		return fmt.Errorf("invalid --type %q, supported values: inclusion, connection", relationType)
	}

	_, orgID, err := resolveIssueScope(ctx, org, project, getIssueBindOrgID, getIssueBindProjectID)
	if err != nil {
		return err
	}
	projectID := ctx.CurrentProject.ProjectID

	parentID := uint64(parent)
	err = bindIssueRelation(ctx, orgID, parentID, &apistructs.IssueRelationCreateRequest{
		IssueID:      parentID,
		RelatedIssue: childIDs,
		ProjectID:    int64(projectID),
		Type:         relationType,
	})
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(issueBindStdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(issueBindOutput{ParentID: parentID, Children: childIDs, Type: relationType})
}

func parseIssueBindChildren(children string) ([]uint64, error) {
	parts := splitIssueListCSV(children)
	if len(parts) == 0 {
		return nil, fmt.Errorf("--children is required")
	}
	result := make([]uint64, 0, len(parts))
	for _, part := range parts {
		childID, err := parsePositiveUint64(part, "--children")
		if err != nil {
			return nil, err
		}
		result = append(result, childID)
	}
	return result, nil
}

func parsePositiveUint64(raw string, flagName string) (uint64, error) {
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q", flagName, raw)
	}
	if value == 0 {
		return 0, fmt.Errorf("%s value must be greater than 0", flagName)
	}
	return value, nil
}
