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

package oas3_test

// timeout
//import (
//	"bytes"
//	"io/ioutil"
//	"testing"
//
//	"github.com/erda-project/erda/pkg/swagger/oas3"
//)
//
//const petstore = "./testdata/petstore-oas3.json"
//
//// 测试 MarshalYaml 序列化结果的一致性
//// 重复执行序列化 100 次, 如果发生两次结果值不一致, 则测试失败
//func TestMarshalYamlConsistency(t *testing.T) {
//	data, err := ioutil.ReadFile(petstore)
//	if err != nil {
//		t.Fatalf("failed to ReadFile: %v", err)
//	}
//
//	v3, err := oas3.LoadFromData(data)
//	if err != nil {
//		t.Fatalf("failed to LoadFromData: %v", err)
//	}
//
//	y, err := oas3.MarshalYaml(v3)
//	if err != nil {
//		t.Fatalf("failed to MarshalYaml: %v", err)
//	}
//
//	for i := 0; i < 100; i++ {
//		y2, err := oas3.MarshalYaml(v3)
//		if err != nil {
//			t.Fatalf("failed to MarshalYaml: %v", err)
//		}
//		if !bytes.Equal(y, y2) {
//			t.Fatalf("y is not equal with y2, index: %v", i)
//		}
//	}
//}
