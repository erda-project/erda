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

package adaptor

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

// AuditReceiver audit接收uc审计对象
type AuditReceiver struct {
	bdl *bundle.Bundle
}

// NewAuditReceiver 初始化审计事件接收者
func NewAuditReceiver(bdl *bundle.Bundle) *AuditReceiver {
	return &AuditReceiver{
		bdl: bdl,
	}
}

// Name ....
func (ar *AuditReceiver) Name() string {
	return "audit_receiver"
}

// SendAudits .....
func (ar *AuditReceiver) SendAudits(ucaudits *apistructs.UCAuditsListResponse) ([]int64, error) {
	if len(ucaudits.Result) == 0 {
		return nil, nil
	}
	batchReq, _, ucIDs := ucaudits.Convert2SysUCAuditBatchCreateRequest()
	logrus.Infof("%v is starting sync %v data", ar.Name(), len(batchReq.Audits))

	if err := ar.bdl.BatchCreateAuditEvent(&batchReq); err != nil {
		return ucIDs, err
	}

	return nil, nil
}
