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

package sheet_issue

import (
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
)

const (
	// Common
	fieldCommon             = "Common"
	fieldID                 = "ID"
	fieldIterationName      = "IterationName"
	fieldIssueType          = "IssueType"
	fieldIssueTitle         = "IssueTitle"
	fieldContent            = "Content"
	fieldState              = "State"
	fieldPriority           = "Priority"
	fieldComplexity         = "Complexity"
	fieldSeverity           = "Severity"
	fieldCreatorName        = "CreatorName"
	fieldAssigneeName       = "AssigneeName"
	fieldCreatedAt          = "CreatedAt"
	fieldPlanStartedAt      = "PlanStartedAt"
	fieldPlanFinishedAt     = "PlanFinishedAt"
	fieldStartAt            = "StartAt"
	fieldFinishAt           = "FinishAt"
	fieldEstimateTime       = "EstimateTime"
	fieldLabels             = "Labels"
	fieldConnectionIssueIDs = "ConnectionIssueIDs"
	fieldCustomFields       = "CustomFields"

	// RequirementOnly
	fieldRequirementOnly   = "RequirementOnly"
	fieldInclusionIssueIDs = "InclusionIssueIDs"

	// TaskOnly
	fieldTaskOnly = "TaskOnly"
	fieldTaskType = "TaskType"

	// BugOnly
	fieldBugOnly     = "BugOnly"
	fieldOwnerName   = "OwnerName"
	fieldSource      = "Source"
	fieldReopenCount = "ReopenCount"

	// Error
	fieldError = "Error"
)

var excelFields = []string{
	// Common
	fieldCommon,
	fieldID,
	fieldIterationName,
	fieldIssueType,
	fieldIssueTitle,
	fieldContent,
	fieldState,
	fieldPriority,
	fieldComplexity,
	fieldSeverity,
	fieldCreatorName,
	fieldAssigneeName,
	fieldCreatedAt,
	fieldPlanStartedAt,
	fieldPlanFinishedAt,
	fieldStartAt,
	fieldFinishAt,
	fieldEstimateTime,
	fieldLabels,
	fieldConnectionIssueIDs,
	fieldCustomFields,

	// RequirementOnly
	fieldRequirementOnly,
	fieldInclusionIssueIDs,

	// TaskOnly
	fieldTaskOnly,
	fieldTaskType,

	// BugOnly
	fieldBugOnly,
	fieldOwnerName,
	fieldSource,
	fieldReopenCount,
}

var (
	i18nMapByText = make(map[string]string)
)

func InitI18nMap(data *vars.DataForFulfill) {
	for _, key := range excelFields {
		i18nMapByText[key] = key // self
		// en-US
		enLang, _ := i18n.ParseLanguageCode("en-US")
		i18nMapByText[data.Tran.Text(enLang, key)] = key
		// zh-CN
		zhLang, _ := i18n.ParseLanguageCode("zh-CN")
		i18nMapByText[data.Tran.Text(zhLang, key)] = key
	}
}
