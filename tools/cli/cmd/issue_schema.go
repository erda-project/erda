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
	"sort"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var ISSUESCHEMA = command.Command{
	ParentName: "ISSUE",
	Name:       "schema",
	ShortHelp:  "show current project issue schema as JSON",
	Example:    "$ erda-cli issue schema --type bug",
	Flags: []command.Flag{
		command.StringFlag{Name: "org", Doc: "organization name; defaults to current workspace context", DefaultValue: ""},
		command.StringFlag{Name: "project", Doc: "project name; defaults to current workspace context", DefaultValue: ""},
		command.StringFlag{Name: "type", Doc: "issue type: requirement, task, bug; empty means all supported types", DefaultValue: ""},
	},
	Run: IssueSchema,
}

var (
	getIssueSchemaOrgID                 = common.GetOrgID
	getIssueSchemaProjectID             = common.GetProjectID
	listIssueSchemaStates               = common.ListState
	listIssueSchemaLabels               = common.ListIssueLabels
	listIssueSchemaProperties           = common.ListIssueProperties
	listIssueSchemaStages               = common.ListIssueStages
	issueSchemaStdout         io.Writer = os.Stdout
)

type issueSchemaDocument struct {
	Project    issueSchemaProject `json:"project"`
	IssueTypes []issueTypeSchema  `json:"issueTypes"`
	Labels     []issueSchemaLabel `json:"labels"`
}

type issueSchemaProject struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type issueTypeSchema struct {
	Type           string             `json:"type"`
	DisplayName    string             `json:"displayName"`
	RequiredFields []string           `json:"requiredFields"`
	Fields         []issueSchemaField `json:"fields"`
	States         []issueSchemaState `json:"states"`
}

type issueSchemaField struct {
	Name        string              `json:"name"`
	DisplayName string              `json:"displayName"`
	ValueType   string              `json:"valueType"`
	Description string              `json:"description,omitempty"`
	Required    bool                `json:"required"`
	Source      string              `json:"source"`
	Options     []issueSchemaOption `json:"options,omitempty"`
}

type issueSchemaOption struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Value       string `json:"value,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Index       int64  `json:"index,omitempty"`
}

type issueSchemaState struct {
	ID          int64                 `json:"id"`
	Name        string                `json:"name"`
	Belong      string                `json:"belong"`
	Transitions []issueSchemaStateRef `json:"transitions"`
}

type issueSchemaStateRef struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Belong string `json:"belong"`
}

type issueSchemaLabel struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	Type      string `json:"type"`
	ProjectID uint64 `json:"projectID"`
}

func IssueSchema(ctx *command.Context, org string, project string, issueType string) error {
	_, orgID, err := resolveIssueScope(ctx, org, project, getIssueSchemaOrgID, getIssueSchemaProjectID)
	if err != nil {
		return err
	}
	projectName := ctx.CurrentProject.Project
	projectID := ctx.CurrentProject.ProjectID

	types, err := parseIssueSchemaTypes(issueType)
	if err != nil {
		return err
	}

	doc := issueSchemaDocument{
		Project: issueSchemaProject{ID: projectID, Name: projectName},
	}
	for _, typ := range types {
		schema, err := buildIssueTypeSchema(ctx, orgID, projectID, typ)
		if err != nil {
			return err
		}
		doc.IssueTypes = append(doc.IssueTypes, schema)
	}

	labels, err := listIssueSchemaLabels(ctx, projectID)
	if err != nil {
		return err
	}
	doc.Labels = toIssueSchemaLabels(labels)

	encoder := json.NewEncoder(issueSchemaStdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}

func parseIssueSchemaTypes(issueType string) ([]apistructs.IssueType, error) {
	if issueType == "" {
		return []apistructs.IssueType{apistructs.IssueTypeRequirement, apistructs.IssueTypeTask, apistructs.IssueTypeBug}, nil
	}

	switch strings.ToLower(issueType) {
	case "requirement":
		return []apistructs.IssueType{apistructs.IssueTypeRequirement}, nil
	case "task":
		return []apistructs.IssueType{apistructs.IssueTypeTask}, nil
	case "bug":
		return []apistructs.IssueType{apistructs.IssueTypeBug}, nil
	default:
		return nil, fmt.Errorf("invalid issue type %q, supported values: requirement, task, bug", issueType)
	}
}

func buildIssueTypeSchema(ctx *command.Context, orgID, projectID uint64, issueType apistructs.IssueType) (issueTypeSchema, error) {
	states, err := listIssueSchemaStates(ctx, orgID, apistructs.IssueStateRelationGetRequest{
		ProjectID: projectID,
		IssueType: issueType,
	})
	if err != nil {
		return issueTypeSchema{}, err
	}
	properties, err := listIssueSchemaProperties(ctx, orgID, projectID, issueType)
	if err != nil {
		return issueTypeSchema{}, err
	}
	stages, err := listIssueSchemaStages(ctx, orgID, issueType)
	if err != nil {
		return issueTypeSchema{}, err
	}

	fields := append(baseIssueSchemaFields(issueType, stages), customIssueSchemaFields(properties)...)
	return issueTypeSchema{
		Type:           string(issueType),
		DisplayName:    issueType.GetZhName(),
		RequiredFields: requiredIssueSchemaFields(fields),
		Fields:         fields,
		States:         toIssueSchemaStates(states),
	}, nil
}

func baseIssueSchemaFields(issueType apistructs.IssueType, stages []apistructs.IssueStage) []issueSchemaField {
	fields := []issueSchemaField{
		requiredBaseField("type", "类型", "string"),
		requiredBaseField("title", "标题", "string"),
		userBaseField("assignee", "处理人", true),
		requiredBaseField("iterationID", "迭代", "uint64"),
		optionalBaseField("content", "内容", "string"),
		optionalBaseField("priority", "优先级", "enum", priorityOptions()),
		optionalBaseField("labels", "标签", "string[]"),
		optionalBaseField("planStartedAt", "计划开始时间", "datetime"),
		optionalBaseField("planFinishedAt", "计划完成时间", "datetime"),
		userBaseField("owner", "负责人", false),
	}

	switch issueType {
	case apistructs.IssueTypeRequirement:
		fields = append(fields, optionalBaseField("complexity", "复杂度", "enum", complexityOptions()))
	case apistructs.IssueTypeTask:
		fields = append(fields,
			optionalBaseField("taskType", "任务类型", "enum", stageOptions(stages)),
			optionalBaseField("manHour", "工时", "object"),
		)
	case apistructs.IssueTypeBug:
		fields = append(fields,
			optionalBaseField("severity", "严重程度", "enum", severityOptions()),
			optionalBaseField("bugStage", "缺陷阶段", "enum", stageOptions(stages)),
			optionalBaseField("manHour", "工时", "object"),
		)
	}
	return fields
}

func requiredBaseField(name, displayName, valueType string) issueSchemaField {
	return issueSchemaField{Name: name, DisplayName: displayName, ValueType: valueType, Required: true, Source: "base"}
}

func optionalBaseField(name, displayName, valueType string, options ...[]issueSchemaOption) issueSchemaField {
	field := issueSchemaField{Name: name, DisplayName: displayName, ValueType: valueType, Source: "base"}
	if len(options) > 0 {
		field.Options = options[0]
	}
	return field
}

func userBaseField(name, displayName string, required bool) issueSchemaField {
	return issueSchemaField{
		Name:        name,
		DisplayName: displayName,
		ValueType:   "user",
		Description: "numeric Erda user ID string, not username or nickname; use `erda-cli whoami` for current UserID",
		Required:    required,
		Source:      "base",
	}
}

func customIssueSchemaFields(properties []apistructs.IssuePropertyIndex) []issueSchemaField {
	fields := make([]issueSchemaField, 0, len(properties))
	for _, property := range properties {
		name := property.PropertyName
		if name == "" {
			continue
		}
		displayName := property.DisplayName
		if displayName == "" {
			displayName = name
		}
		fields = append(fields, issueSchemaField{
			Name:        name,
			DisplayName: displayName,
			ValueType:   string(property.PropertyType),
			Required:    property.Required,
			Source:      "custom",
			Options:     propertyOptions(property.EnumeratedValues),
		})
	}
	return fields
}

func requiredIssueSchemaFields(fields []issueSchemaField) []string {
	required := make([]string, 0)
	for _, field := range fields {
		if field.Required {
			required = append(required, field.Name)
		}
	}
	return required
}

func toIssueSchemaStates(states []apistructs.IssueStateRelation) []issueSchemaState {
	stateRefs := make(map[int64]issueSchemaStateRef, len(states))
	for _, state := range states {
		stateRefs[state.StateID] = issueSchemaStateRef{
			ID:     state.StateID,
			Name:   state.StateName,
			Belong: string(state.StateBelong),
		}
	}

	result := make([]issueSchemaState, 0, len(states))
	for _, state := range states {
		transitions := make([]issueSchemaStateRef, 0, len(state.StateRelation))
		for _, targetID := range state.StateRelation {
			if target, ok := stateRefs[targetID]; ok {
				transitions = append(transitions, target)
			}
		}
		result = append(result, issueSchemaState{
			ID:          state.StateID,
			Name:        state.StateName,
			Belong:      string(state.StateBelong),
			Transitions: transitions,
		})
	}
	return result
}

func toIssueSchemaLabels(labels []apistructs.ProjectLabel) []issueSchemaLabel {
	result := make([]issueSchemaLabel, 0, len(labels))
	for _, label := range labels {
		result = append(result, issueSchemaLabel{
			ID:        label.ID,
			Name:      label.Name,
			Color:     label.Color,
			Type:      string(label.Type),
			ProjectID: label.ProjectID,
		})
	}
	return result
}

func propertyOptions(values []apistructs.Enumerate) []issueSchemaOption {
	options := make([]issueSchemaOption, 0, len(values))
	for _, value := range values {
		options = append(options, issueSchemaOption{
			ID:          value.ID,
			Name:        value.Name,
			Value:       value.Name,
			DisplayName: value.Name,
			Index:       value.Index,
		})
	}
	sort.SliceStable(options, func(i, j int) bool {
		return options[i].Index < options[j].Index
	})
	return options
}

func priorityOptions() []issueSchemaOption {
	options := make([]issueSchemaOption, 0, len(apistructs.IssuePriorityList))
	for _, priority := range apistructs.IssuePriorityList {
		options = append(options, enumOption(string(priority), priority.GetZhName()))
	}
	return options
}

func complexityOptions() []issueSchemaOption {
	complexities := []apistructs.IssueComplexity{
		apistructs.IssueComplexityHard,
		apistructs.IssueComplexityNormal,
		apistructs.IssueComplexityEasy,
	}
	options := make([]issueSchemaOption, 0, len(complexities))
	for _, complexity := range complexities {
		options = append(options, enumOption(string(complexity), complexity.GetZhName()))
	}
	return options
}

func severityOptions() []issueSchemaOption {
	options := make([]issueSchemaOption, 0, len(apistructs.IssueSeveritys))
	for _, severity := range apistructs.IssueSeveritys {
		options = append(options, enumOption(string(severity), severity.GetZhName()))
	}
	return options
}

func stageOptions(stages []apistructs.IssueStage) []issueSchemaOption {
	options := make([]issueSchemaOption, 0, len(stages))
	for _, stage := range stages {
		value := stage.Value
		if value == "" {
			value = stage.Name
		}
		options = append(options, issueSchemaOption{
			ID:          stage.ID,
			Name:        stage.Name,
			Value:       value,
			DisplayName: stage.Name,
		})
	}
	return options
}

func enumOption(value, displayName string) issueSchemaOption {
	return issueSchemaOption{Name: value, Value: value, DisplayName: displayName}
}
