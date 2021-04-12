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

package bundle

//import (
//	"os"
//	"strconv"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/require"
//
//	"github.com/erda-project/erda/apistructs"
//)
//
//func TestBundle_CreateEvent(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.marathon.l4lb.thisdcos.directory:9093")
//	defer func() {
//		os.Unsetenv("CMDB_ADDR")
//	}()
//	b := New(WithCMDB())
//	tm := strconv.FormatInt(time.Now().Unix(), 10)
//	audit := apistructs.Audit{
//		UserID:       "1000008",
//		ScopeType:    "app",
//		ScopeID:      2,
//		AppID:        4,
//		ProjectID:    2,
//		OrgID:        2,
//		Context:      map[string]interface{}{"appId": "4", "projectId": "2"},
//		TemplateName: "createApp",
//		AuditLevel:   "p3",
//		Result:       "success",
//		ErrorMsg:     "",
//		StartTime:    tm,
//		EndTime:      tm,
//		ClientIP:     "1.1.1.1",
//		UserAgent:    "chrom",
//	}
//	err := b.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: audit})
//	require.NoError(t, err)
//}
