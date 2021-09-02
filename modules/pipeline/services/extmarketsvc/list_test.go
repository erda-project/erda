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
