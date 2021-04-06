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

package apistructs

import (
	"time"
)

type TPRecord struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	ApplicationID   int64             `json:"applicationId"`
	ProjectID       int64             `json:"projectId"`
	BuildID         int64             `json:"buildId"`
	Name            string            `json:"name"`
	UUID            string            `json:"uuid"`
	ApplicationName string            `json:"applicationName"`
	Output          string            `json:"output"`
	Desc            string            `json:"desc"`
	OperatorID      string            `json:"operatorId"`
	OperatorName    string            `json:"operatorName"`
	CommitID        string            `json:"commitId"`
	Branch          string            `json:"branch"`
	GitRepo         string            `json:"gitRepo"`
	CaseDir         string            `json:"caseDir"`
	Application     string            `json:"application"`
	Avatar          string            `json:"avatar,omitempty"`
	TType           TestType          `json:"type"`
	Totals          *TestTotals       `json:"totals"`
	ParserType      string            `json:"parserType"`
	Extra           map[string]string `json:"extra,omitempty"`
	Envs            map[string]string `json:"envs"`
	Workspace       DiceWorkspace     `json:"workspace"`
	Suites          []*TestSuite      `json:"suites"`
}
