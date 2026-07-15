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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var ISSUECREATE = command.Command{
	ParentName: "ISSUE",
	Name:       "create",
	ShortHelp:  "create an issue from JSON file",
	Example:    "$ erda-cli issue create --file issue.json",
	Flags: []command.Flag{
		command.StringFlag{Name: "org", Doc: "organization name; defaults to current workspace context", DefaultValue: ""},
		command.StringFlag{Name: "project", Doc: "project name; defaults to current workspace context", DefaultValue: ""},
		command.StringFlag{Name: "file", Doc: "issue JSON file generated from issue schema", DefaultValue: ""},
		command.StringFlag{Name: "json", Doc: "inline issue JSON, mutually exclusive with --file", DefaultValue: ""},
	},
	Run: IssueCreate,
}

var (
	getIssueCreateOrgID                   = common.GetOrgID
	getIssueCreateProjectID               = common.GetProjectID
	listIssueCreateProperties             = common.ListIssueProperties
	listIssueCreateStages                 = common.ListIssueStages
	createIssue                           = common.CreateIssue
	createIssuePropertyInstance           = common.CreateIssuePropertyInstance
	createIssueRelation                   = common.CreateIssueRelation
	issueCreateStdout           io.Writer = os.Stdout
)

type issueCreateFile struct {
	apistructs.IssueCreateRequest
	CustomFields map[string]json.RawMessage `json:"customFields"`
	ParentID     uint64                     `json:"parentID"`
	RelationType string                     `json:"relationType"`
}

type issueCreateOutput struct {
	ID uint64 `json:"id"`
}

func IssueCreate(ctx *command.Context, org string, project string, file string, inlineJSON string) error {
	if (file == "") == (inlineJSON == "") {
		return fmt.Errorf("exactly one of --file or --json is required")
	}

	createFile, err := readIssueCreateInput(file, inlineJSON)
	if err != nil {
		return err
	}

	_, orgID, err := resolveIssueScope(ctx, org, project, getIssueCreateOrgID, getIssueCreateProjectID)
	if err != nil {
		return err
	}
	projectID := ctx.CurrentProject.ProjectID

	req := createFile.IssueCreateRequest
	req.ProjectID = projectID

	properties, err := listIssueCreateProperties(ctx, orgID, projectID, req.Type)
	if err != nil {
		return err
	}
	stages, err := listIssueCreateStages(ctx, orgID, req.Type)
	if err != nil {
		return err
	}
	if err := normalizeIssueStage(&req, stages); err != nil {
		return err
	}
	propertyInstances, err := buildIssuePropertyInstances(orgID, projectID, req.Type, properties, createFile.CustomFields)
	if err != nil {
		return err
	}

	issueID, err := createIssue(ctx, orgID, &req)
	if err != nil {
		return err
	}

	if len(propertyInstances) > 0 {
		err = createIssuePropertyInstance(ctx, &common.IssuePropertyInstanceCreateRequest{
			OrgID:     int64(orgID),
			ProjectID: int64(projectID),
			IssueID:   int64(issueID),
			Property:  propertyInstances,
		})
		if err != nil {
			return err
		}
	}
	if createFile.ParentID != 0 {
		relationType := createFile.RelationType
		if relationType == "" {
			relationType = apistructs.IssueRelationInclusion
		}
		err = createIssueRelation(ctx, orgID, createFile.ParentID, &apistructs.IssueRelationCreateRequest{
			IssueID:      createFile.ParentID,
			RelatedIssue: []uint64{issueID},
			ProjectID:    int64(projectID),
			Type:         relationType,
		})
		if err != nil {
			return err
		}
	}

	encoder := json.NewEncoder(issueCreateStdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(issueCreateOutput{ID: issueID})
}

func readIssueCreateInput(file string, inlineJSON string) (*issueCreateFile, error) {
	content := []byte(inlineJSON)
	if file != "" {
		fileContent, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		content = fileContent
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()

	var createFile issueCreateFile
	if err := decoder.Decode(&createFile); err != nil {
		return nil, err
	}
	if createFile.Type == "" {
		return nil, fmt.Errorf("type is required")
	}
	types, err := parseIssueSchemaTypes(string(createFile.Type))
	if err != nil {
		return nil, err
	}
	createFile.Type = types[0]
	if createFile.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if createFile.IterationID == 0 {
		return nil, fmt.Errorf("iterationID is required, use -1 when the issue does not belong to any iteration")
	}
	if createFile.Assignee == "" && createFile.Type != apistructs.IssueTypeTicket {
		return nil, fmt.Errorf("assignee is required")
	}
	if err := validateIssueCreateUserIDs(&createFile); err != nil {
		return nil, err
	}
	return &createFile, nil
}

func validateIssueCreateUserIDs(createFile *issueCreateFile) error {
	userFields := map[string]string{
		"assignee": createFile.Assignee,
		"creator":  createFile.Creator,
		"owner":    createFile.Owner,
	}
	for field, value := range userFields {
		if err := validateIssueCreateUserID(field, value); err != nil {
			return err
		}
	}
	for i, subscriber := range createFile.Subscribers {
		if err := validateIssueCreateUserID(fmt.Sprintf("subscribers[%d]", i), subscriber); err != nil {
			return err
		}
	}
	return nil
}

func validateIssueCreateUserID(field string, value string) error {
	if value == "" {
		return nil
	}
	if _, err := strconv.ParseUint(value, 10, 64); err != nil {
		return fmt.Errorf("%s must be a numeric Erda user ID, got %q; use erda-cli whoami to get current UserID", field, value)
	}
	return nil
}

func normalizeIssueStage(req *apistructs.IssueCreateRequest, stages []apistructs.IssueStage) error {
	switch req.Type {
	case apistructs.IssueTypeTask:
		if req.TaskType == "" {
			return nil
		}
		value, err := normalizeIssueStageValue(stages, req.TaskType, "taskType")
		if err != nil {
			return err
		}
		req.TaskType = value
	case apistructs.IssueTypeBug:
		if req.BugStage == "" {
			return nil
		}
		value, err := normalizeIssueStageValue(stages, req.BugStage, "bugStage")
		if err != nil {
			return err
		}
		req.BugStage = value
	}
	return nil
}

func normalizeIssueStageValue(stages []apistructs.IssueStage, input string, fieldName string) (string, error) {
	for _, stage := range stages {
		if stage.Value == "" {
			stage.Value = stage.Name
		}
		if input == stage.Value || input == stage.Name {
			return stage.Value, nil
		}
	}
	return "", fmt.Errorf("invalid %s %q, use one of current project stage values or names", fieldName, input)
}

func buildIssuePropertyInstances(orgID, projectID uint64, issueType apistructs.IssueType, properties []apistructs.IssuePropertyIndex, customFields map[string]json.RawMessage) ([]common.IssuePropertyInstance, error) {
	propertiesByName := make(map[string]apistructs.IssuePropertyIndex, len(properties))
	for _, property := range properties {
		if property.PropertyIssueType != "" && property.PropertyIssueType != issuePropertyType(issueType) {
			continue
		}
		propertiesByName[property.PropertyName] = property
	}

	for _, property := range propertiesByName {
		if property.Required && !hasCustomFieldValue(customFields, property.PropertyName) {
			return nil, fmt.Errorf("custom field %q is required", property.PropertyName)
		}
	}

	instances := make([]common.IssuePropertyInstance, 0, len(customFields))
	for name, value := range customFields {
		property, ok := propertiesByName[name]
		if !ok {
			return nil, fmt.Errorf("unknown custom field %q", name)
		}
		instance, err := buildIssuePropertyInstance(orgID, projectID, property, value)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func hasCustomFieldValue(customFields map[string]json.RawMessage, name string) bool {
	value, ok := customFields[name]
	if !ok {
		return false
	}
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 || string(trimmed) == "null" || string(trimmed) == `""` || string(trimmed) == "[]" {
		return false
	}
	return true
}

func buildIssuePropertyInstance(orgID, projectID uint64, property apistructs.IssuePropertyIndex, value json.RawMessage) (common.IssuePropertyInstance, error) {
	instance := common.IssuePropertyInstance{
		PropertyID:        property.PropertyID,
		ScopeID:           property.ScopeID,
		ScopeType:         property.ScopeType,
		OrgID:             int64(orgID),
		PropertyName:      property.PropertyName,
		DisplayName:       property.DisplayName,
		PropertyType:      property.PropertyType,
		Required:          property.Required,
		PropertyIssueType: property.PropertyIssueType,
		Relation:          property.Relation,
		Index:             property.Index,
		EnumeratedValues:  property.EnumeratedValues,
	}
	if instance.ScopeID == 0 {
		instance.ScopeID = int64(projectID)
	}
	if instance.ScopeType == "" {
		instance.ScopeType = apistructs.ProjectScope
	}

	if property.PropertyType.IsOptions() {
		values, err := issuePropertyOptionValues(property, value)
		if err != nil {
			return common.IssuePropertyInstance{}, err
		}
		instance.Values = values
		return instance, nil
	}

	arbitraryValue, err := issuePropertyArbitraryValue(value)
	if err != nil {
		return common.IssuePropertyInstance{}, err
	}
	if property.PropertyType == apistructs.PropertyTypePerson {
		userID, ok := arbitraryValue.(string)
		if !ok {
			return common.IssuePropertyInstance{}, fmt.Errorf("custom field %q must be a numeric Erda user ID string", property.PropertyName)
		}
		if err := validateIssueCreateUserID(fmt.Sprintf("customFields.%s", property.PropertyName), userID); err != nil {
			return common.IssuePropertyInstance{}, err
		}
	}
	instance.ArbitraryValue = arbitraryValue
	return instance, nil
}

func issuePropertyOptionValues(property apistructs.IssuePropertyIndex, value json.RawMessage) ([]int64, error) {
	var rawValues []json.RawMessage
	if bytes.HasPrefix(bytes.TrimSpace(value), []byte("[")) {
		if err := json.Unmarshal(value, &rawValues); err != nil {
			return nil, err
		}
	} else {
		rawValues = []json.RawMessage{value}
	}

	values := make([]int64, 0, len(rawValues))
	for _, rawValue := range rawValues {
		valueID, err := issuePropertyOptionID(property, rawValue)
		if err != nil {
			return nil, err
		}
		values = append(values, valueID)
	}
	return values, nil
}

func issuePropertyOptionID(property apistructs.IssuePropertyIndex, value json.RawMessage) (int64, error) {
	var name string
	if err := json.Unmarshal(value, &name); err == nil {
		for _, option := range property.EnumeratedValues {
			if option.Name == name {
				return option.ID, nil
			}
		}
		return 0, fmt.Errorf("invalid value %q for custom field %q", name, property.PropertyName)
	}

	var id int64
	if err := json.Unmarshal(value, &id); err == nil {
		for _, option := range property.EnumeratedValues {
			if option.ID == id {
				return id, nil
			}
		}
		return 0, fmt.Errorf("invalid value %d for custom field %q", id, property.PropertyName)
	}
	return 0, fmt.Errorf("invalid option value for custom field %q", property.PropertyName)
}

func issuePropertyArbitraryValue(value json.RawMessage) (interface{}, error) {
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()

	var arbitraryValue interface{}
	if err := decoder.Decode(&arbitraryValue); err != nil {
		return nil, err
	}
	return arbitraryValue, nil
}

func issuePropertyType(issueType apistructs.IssueType) apistructs.PropertyIssueType {
	switch issueType {
	case apistructs.IssueTypeRequirement:
		return apistructs.PropertyIssueTypeRequirement
	case apistructs.IssueTypeTask:
		return apistructs.PropertyIssueTypeTask
	case apistructs.IssueTypeBug:
		return apistructs.PropertyIssueTypeBug
	case apistructs.IssueTypeEpic:
		return apistructs.PropertyIssueTypeEpic
	default:
		return apistructs.PropertyIssueType(issueType)
	}
}
