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

package main

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/apis"
)

// use this test function to generate openapiv1 protobuf files.
func TestGenerateProto(t *testing.T) {
	generateProto()
}

func Test_getCompAndSubDirsByAPIName(t *testing.T) {
	apiName := "dop.ADMIN_APPROVE_GET"
	compName, subDirs := getCompAndSubDirsByAPIName(apiName)
	assert.Equal(t, "dop", compName)
	assert.Equal(t, 0, len(subDirs))

	apiName = "testplatform_autotest.ACTION_LOG_GET"
	compName, subDirs = getCompAndSubDirsByAPIName(apiName)
	assert.Equal(t, "testplatform", compName)
	assert.Equal(t, 1, len(subDirs))
	assert.Equal(t, "autotest", subDirs[0])
}

func Test_getAPIVarNameByAPIName(t *testing.T) {
	varName := getAPIVarNameByAPIName("dop.ADMIN_APPROVE_GET")
	assert.Equal(t, "ADMIN_APPROVE_GET", varName)

	varName = getAPIVarNameByAPIName("testplatform_autotest.ACTION_LOG_GET")
	assert.Equal(t, "ACTION_LOG_GET", varName)
}

func Test_lowerFirstChar(t *testing.T) {
	s := ""
	assert.Equal(t, "", lowerFirstChar(s))

	s = "CreatedAt"
	assert.Equal(t, "createdAt", lowerFirstChar(s))
}

func TestParseType(t *testing.T) {
	labels := struct {
		Label map[string]string
	}{}
	rt := reflect.TypeOf(labels)
	f := rt.Field(0)
	assert.Equal(t, reflect.Map, f.Type.Kind())
	assert.Equal(t, reflect.String, f.Type.Key().Kind())
	assert.Equal(t, reflect.String, f.Type.Elem().Kind())

	fmt.Println(f.Type.Key().Kind())

	kt := f.Type.Key()
	fmt.Println(kt.Kind())
	assert.Equal(t, reflect.String, kt.Kind())
}

func TestCreateEmbedMessage(t *testing.T) {
	messages := make(map[string]*Message)
	createEmbedMessage(messages, reflect.TypeOf(apistructs.PipelineCreateRequestV2{}))
	var buff bytes.Buffer
	writeMessages(&buff, messages)
	fmt.Println(buff.String())
}

func TestKind(t *testing.T) {
	type TriggerMode string
	type Data struct {
		Key   string
		Value string
	}
	s := struct {
		apistructs.Header
		apistructs.UserInfoHeader
		ID        uint64
		CronID    *uint64
		CreatedAt time.Time
		UpdatedAt *time.Time
		Interface interface{}
		Mode      TriggerMode
		Data      Data `json:"data"`
	}{}
	messages := make(map[string]*Message)
	rt := reflect.TypeOf(s)
	fmt.Printf("%-15s %-30s %-15s %-15s\n", "f.name", "f.type", "f.kind", "proto.type")
	fmt.Printf("%-15s %-30s %-15s %-15s\n", "------", "------", "------", "------")
	for i := 0; i < rt.NumField(); i++ {
		rf := rt.Field(i)
		f := makeMessageField(messages, rf)
		if f != nil {
			fmt.Printf("%-15s %-30s %-15s %-15s\n", rf.Name, rf.Type, rf.Type.Kind(), f.ProtoType)
		}
	}

	fmt.Printf("\npolished response:\n\n")

	wAPI := WrappedApiSpec{API: apis.ApiSpec{ResponseType: s}}
	//polishResponseTypeBeforeParse(&wAPI)
	messages = make(map[string]*Message)
	rt = reflect.TypeOf(wAPI.API.ResponseType)
	fmt.Printf("%-15s %-30s %-15s %-15s\n", "f.name", "f.type", "f.kind", "proto.type")
	fmt.Printf("%-15s %-30s %-15s %-15s\n", "------", "------", "------", "------")
	for i := 0; i < rt.NumField(); i++ {
		rf := rt.Field(i)
		f := makeMessageField(messages, rf)
		if f != nil {
			fmt.Printf("%-15s %-30s %-15s %-15s\n", rf.Name, rf.Type, rf.Type.Kind(), f.ProtoType)
		}
	}
}

func TestParse(t *testing.T) {
	messages := make(map[string]*Message)
	createEmbedMessage(messages, reflect.TypeOf(apistructs.PagingProjectDTO{}))
	for k, v := range messages {
		fmt.Println(k, v)
	}
}

func TestFindIncorrectApiSpec(t *testing.T) {
	pathVarReg := regexp.MustCompile(`<([^<>]+)>`)
	for i, API := range APIs {
		matches := pathVarReg.FindAllStringSubmatch(API.Path, -1)
		if len(matches) == 0 {
			continue
		}
		rt := reflect.TypeOf(API.RequestType)
		if rt == nil {
			logrus.Infof("missing var, APIName: %s", APINames[i])
			continue
		}
		if rt.Kind() != reflect.Struct {
			logrus.Infof("missing var, APIName: %s", APINames[i])
			continue
		}
		allFieldNames := make(map[string]struct{})
		for i := 0; i < rt.NumField(); i++ {
			allFieldNames[rt.Field(i).Name] = struct{}{}
		}
		for _, match := range matches {
			v := match[1]
			if _, ok := allFieldNames[v]; !ok {
				logrus.Infof("var: %s not found in request: %s, APIName: %s", v, rt.String(), APINames[i])
			}
		}
	}
}

func TestRegexp(t *testing.T) {
	pathVarReg := regexp.MustCompile(`{([^{}]+)}`)
	s := "/app/{applicationId}/{workspace}"
	fmt.Println(pathVarReg.FindAllStringSubmatch(s, -1))
}

func Test_getProtoPackageByAPIName(t *testing.T) {
	pkg := getProtoPackageByAPIName("dop.TestAPI")
	assert.Equal(t, "erda.openapiv1.dop", pkg)
}

func Test_getJsonTagValue(t *testing.T) {
	s := struct {
		A string `json:"a"`
		B string `path:"b"`
		C string `json:"cc" path:"c"`
		D string `other:"d"`
	}{}
	rt := reflect.TypeOf(s)
	fa, ok := rt.FieldByName("A")
	assert.True(t, ok)
	assert.Equal(t, "a", getJsonTagValue(fa.Tag))
	fb, ok := rt.FieldByName("B")
	assert.True(t, ok)
	assert.Equal(t, "b", getJsonTagValue(fb.Tag))
	fc, ok := rt.FieldByName("C")
	assert.True(t, ok)
	assert.Equal(t, "cc", getJsonTagValue(fc.Tag))
	fd, ok := rt.FieldByName("D")
	assert.True(t, ok)
	assert.Equal(t, "", getJsonTagValue(fd.Tag))
}

func Test_getCompAndSubDirsByAPIName1(t *testing.T) {
	type args struct {
		apiName string
	}
	tests := []struct {
		name         string
		args         args
		wantCompName string
		wantSubDirs  []string
	}{
		{
			name:         "msp_apm",
			args:         args{apiName: "msp_apm"},
			wantCompName: "msp",
			wantSubDirs:  []string{"apm"},
		},
		{
			name:         "dop",
			args:         args{apiName: "dop"},
			wantCompName: "dop",
			wantSubDirs:  nil,
		},
		{
			name:         "openapi_component_protocol",
			args:         args{apiName: "openapi_component_protocol"},
			wantCompName: "openapi",
			wantSubDirs:  []string{"component", "protocol"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCompName, gotSubDirs := getCompAndSubDirsByAPIName(tt.args.apiName)
			assert.Equalf(t, tt.wantCompName, gotCompName, "getCompAndSubDirsByAPIName(%v)", tt.args.apiName)
			assert.Equalf(t, tt.wantSubDirs, gotSubDirs, "getCompAndSubDirsByAPIName(%v)", tt.args.apiName)
		})
	}
}
