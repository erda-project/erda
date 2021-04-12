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

package dbclient

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/qaparser/types"
)

const (
	PageNo         = "pageNo"
	PageSize       = "pageSize"
	ExtraEndpoint  = "extra_cs_endpoint"
	ExtraAccessKey = "extra_cs_ak"
	ExtraSecretKey = "extra_cs_sk"
	ExtraBucket    = "extra_cs_bucket"

	HTTP = "HTTP"
)

var taskStepsNames = []string{"repo", "test"}

type TPRecordDO struct {
	ID        uint64    `xorm:"pk autoincr 'id'" json:"id"`
	CreatedAt time.Time `xorm:"created" json:"createdAt"`
	UpdatedAt time.Time `xorm:"updated" json:"updatedAt"`

	ApplicationID   int64                    `xorm:"app_id" json:"applicationId"`
	ProjectID       int64                    `xorm:"project_id" json:"projectId" validate:"required"`
	BuildID         int64                    `xorm:"build_id" json:"buildId"`
	Name            string                   `xorm:"name" json:"name" validate:"required"`
	UUID            string                   `xorm:"uuid" json:"uuid"`
	ApplicationName string                   `xorm:"app_name" json:"applicationName"`
	Output          string                   `xorm:"output" json:"output"`
	Desc            string                   `xorm:"desc" json:"desc"`
	OperatorID      string                   `xorm:"operator_id" json:"operatorId"`
	OperatorName    string                   `xorm:"operator_name" json:"operatorName"`
	CommitID        string                   `xorm:"commit_id" json:"commitId"`
	Branch          string                   `xorm:"branch" json:"branch" validate:"required"`
	GitRepo         string                   `xorm:"git_repo" json:"gitRepo" validate:"required"`
	CaseDir         string                   `xorm:"case_dir" json:"caseDir"`
	Application     string                   `xorm:"-" json:"application"`
	Avatar          string                   `xorm:"-" json:"avatar,omitempty"`
	TType           apistructs.TestType      `xorm:"type" json:"type" validate:"required"`
	Totals          *apistructs.TestTotals   `xorm:"totals json" json:"totals"`
	ParserType      types.TestParserType     `xorm:"parser_type" json:"parserType"`
	Extra           map[string]string        `xorm:"varchar(1024) 'extra'" json:"extra,omitempty"`
	Envs            map[string]string        `xorm:"varchar(1024) 'envs'" json:"envs"`
	Workspace       apistructs.DiceWorkspace `xorm:"workspace" json:"workspace" validate:"required"`
	Suites          []*apistructs.TestSuite  `xorm:"longtext 'suites'" json:"suites"`
}

func (TPRecordDO) TableName() string {
	return "qa_test_records"
}

func NewTPRecordDO() *TPRecordDO {
	return &TPRecordDO{
		Totals: &apistructs.TestTotals{
			Statuses: make(map[apistructs.TestStatus]int),
		},
		Extra: make(map[string]string),
	}
}

func (tp *TPRecordDO) SetCols(cols interface{}) (*TPRecordDO, error) {
	bytes, err := json.Marshal(cols)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("json:%s", string(bytes))

	if err := json.Unmarshal(bytes, tp); err != nil {
		return nil, err
	}

	return tp, nil
}

func (tp *TPRecordDO) SetTPType(tpType apistructs.TestType) *TPRecordDO {
	tp.TType = tpType
	return tp
}

func (tp *TPRecordDO) GetExtraInfo(key string) string {
	if tp.Extra == nil {
		return ""
	}
	return tp.Extra[key]
}

func (tp *TPRecordDO) GetRecordName() string {
	// date-type
	return fmt.Sprintf("%s-%s", tp.CreatedAt.Format("2006-01-02-15-04-05"), tp.TType)
}

func (tp *TPRecordDO) GenSubject() string {
	// TODO get more info in email subject
	return "[TP] 测试报告"
}

func (tp *TPRecordDO) EraseSensitiveInfo() {
	if tp.Extra != nil {
		delete(tp.Extra, ExtraEndpoint)
		delete(tp.Extra, ExtraSecretKey)
		delete(tp.Extra, ExtraAccessKey)
		delete(tp.Extra, ExtraBucket)
	}
}
