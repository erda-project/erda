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
//	"time"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/apistructs"
//	"github.com/erda-project/erda/pkg/httpclient"
//)
//
//func TestBundle_ListGittarFileTreeNodes(t *testing.T) {
//	hc := httpclient.New(httpclient.WithEnableAutoRetry(false))
//	os.Setenv("GITTAR_ADAPTOR_ADDR", "gittar-adaptor.default.svc.cluster.local:1086")
//	bdl := New(WithGittarAdaptor(), WithHTTPClient(hc))
//	begin := time.Now()
//	nodes, err := bdl.ListGittarFileTreeNodes(apistructs.UnifiedFileTreeNodeListRequest{
//		Scope:        "project-app",
//		ScopeID:      "59",
//		Pinode:       "NTkvMjI=",
//		IdentityInfo: apistructs.IdentityInfo{UserID: "1"},
//	}, 1)
//	end := time.Now()
//	fmt.Println(end.Sub(begin))
//	assert.NoError(t, err)
//	fmt.Println(len(nodes))
//}
