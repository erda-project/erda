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
//	"fmt"
//	"os"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/apistructs"
//)
//
//func TestBundle_GetNotifyConfigMS(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	defer func() {
//		os.Unsetenv("CMDB_ADDR")
//	}()
//	bdl := New(WithCMDB())
//	cfg, err := bdl.GetNotifyConfigMS("2", "1")
//	assert.NoError(t, err)
//	assert.Equal(t, false, cfg)
//}
//
//func TestBundle_NotifyList(t *testing.T) {
//	os.Setenv("MONITOR_ADDR", "monitor.default.svc.cluster.local:7096")
//	defer func() {
//		os.Unsetenv("MONITOR_ADDR")
//	}()
//	bdl := New(WithMonitor())
//	req := apistructs.NotifyPageRequest{
//		ScopeId: "18",
//		Scope:   "app",
//		UserId:  "2",
//		OrgId:   "1",
//	}
//	cfg, err := bdl.NotifyList(req)
//	assert.NoError(t, err)
//	fmt.Printf("%+v", cfg)
//}
