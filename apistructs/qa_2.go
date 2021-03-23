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
