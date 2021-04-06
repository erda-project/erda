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

package extmarketsvc

//import (
//	"encoding/json"
//	"fmt"
//	"log"
//	"os"
//	"testing"
//
//	"github.com/erda-project/erda/bundle"
//)
//
//func TestExtMarketSvc_SearchActions(t *testing.T) {
//	os.Setenv("DICEHUB_ADDR", "dicehub.default.svc.cluster.local:10000")
//	bdl := bundle.New(bundle.WithDiceHub())
//	s := New(bdl)
//	m, n, err := s.SearchActions([]string{"java-sec2"})
//	if err != nil {
//		log.Fatalln(err)
//	}
//	for _, v := range m {
//		b, _ := json.MarshalIndent(&v, "", "  ")
//		fmt.Println(string(b))
//	}
//	for _, v := range n {
//		b, _ := json.MarshalIndent(&v, "", "  ")
//		fmt.Println(string(b))
//	}
//}
