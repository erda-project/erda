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

package filter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"bou.ke/monkey"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/protobuf/proto-go/cp/pb"
	i18n2 "github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/cmp"
	"github.com/erda-project/erda/internal/apps/cmp/component-protocol/cputil"
)

func getTestURLQuery() (Values, string) {
	v := Values{
		Kind:      []string{"test"},
		Namespace: "test",
		Status:    []string{"test"},
		Search:    "test",
	}
	data, _ := json.Marshal(v)
	encode := base64.StdEncoding.EncodeToString(data)
	return v, encode
}

func isEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestComponentFilter_DecodeURLQuery(t *testing.T) {
	values, res := getTestURLQuery()
	filter := &ComponentFilter{
		sdk: &cptype.SDK{
			InParams: map[string]interface{}{
				"filter__urlQuery": res,
			},
		},
	}
	if err := filter.DecodeURLQuery(); err != nil {
		t.Errorf("test failed, %v", err)
	}
	if filter.State.Values.Namespace != values.Namespace || !isEqual(filter.State.Values.Status, values.Status) ||
		!isEqual(filter.State.Values.Kind, values.Kind) || filter.State.Values.Search != values.Search {
		t.Errorf("test failed, edcode result is not expected")
	}
}

func TestComponentFilter_GenComponentState(t *testing.T) {
	component := &cptype.Component{
		State: map[string]interface{}{
			"clusterName": "test",
			"conditions": []Condition{
				{
					Key:         "test",
					Label:       "test",
					Placeholder: "test",
					Type:        "test",
					Fixed:       true,
					Options: []Option{
						{
							Label: "test",
							Value: "test",
						},
					},
				},
			},
			"values": Values{
				Kind:      []string{"test"},
				Namespace: "test",
				Status:    []string{"test"},
				Search:    "test",
			},
		},
	}
	src, err := json.Marshal(component.State)
	if err != nil {
		t.Errorf("test failed, %v", err)
	}

	f := &ComponentFilter{}
	if err := f.GenComponentState(component); err != nil {
		t.Errorf("test failed, %v", err)
	}

	dst, err := json.Marshal(f.State)
	if err != nil {
		t.Errorf("test failed, %v", err)
	}

	if string(src) != string(dst) {
		t.Error("test failed, generate result is unexpected")
	}
}

func TestComponentFilter_EncodeURLQuery(t *testing.T) {
	values, data := getTestURLQuery()
	f := ComponentFilter{
		State: State{
			Values: values,
		},
	}
	if err := f.EncodeURLQuery(); err != nil {
		t.Errorf("test failed, %v", err)
	}
	if f.State.FilterURLQuery != data {
		t.Error("test failed, encode url query result is unexpected")
	}
}

func TestComponentFilter_Transfer(t *testing.T) {
	component := &ComponentFilter{
		State: State{
			ClusterName: "testClusterName",
			Conditions: []Condition{
				{
					HaveFilter:  true,
					Key:         "testKey",
					Label:       "testLabel",
					Placeholder: "testPlaceHolder",
					Type:        "testType",
					Fixed:       true,
					Options: []Option{
						{
							Label: "testLabel",
							Value: "testValue",
						},
					},
				},
			},
			Values: Values{
				Kind:      []string{"test"},
				Namespace: "test",
				Status:    []string{"test"},
				Node:      []string{"node"},
				Search:    "test",
			},
			FilterURLQuery: "testURLQuery",
		},
		Operations: map[string]interface{}{
			"testOp": Operation{
				Key:    "testKey",
				Reload: true,
			},
		},
	}
	c := &cptype.Component{}
	component.Transfer(c)
	ok, err := cputil.IsDeepEqual(c, component)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Errorf("test failed, json is not equal")
	}
}

func TestHasSuffix(t *testing.T) {
	dev := "project-1-dev"
	_, ok := hasSuffix(dev)
	if !ok {
		t.Error("test failed, expected to have suffix \"-dev\", actual not")
	}

	stage := "project-2-staging"
	_, ok = hasSuffix(stage)
	if !ok {
		t.Error("test failed, expected to have suffix \"-staging\", actual not")
	}

	test := "project-3-test"
	_, ok = hasSuffix(test)
	if !ok {
		t.Error("test failed, expected to have suffix \"-test\", actual not")
	}

	prod := "project-4-prod"
	_, ok = hasSuffix(prod)
	if !ok {
		t.Error("test failed, expected to have suffix \"-prod\", actual not")
	}

	other := "default"
	_, ok = hasSuffix(other)
	if ok {
		t.Error("test failed, expected to not have suffix, actual do")
	}
}

// Mocking the GetAllNamespacesFromCache function
func mockGetAllNamespacesFromCache(context.Context, cmp.SteveServer, string, string, string) ([]string, error) {
	return []string{
		"project-test", "project-dev", "project-x-prod",
		"project-1-dev", "project-1-test", "project-1-staging", "project-2-prod",
		"group-addon-kafka--va2f089955c034f91831fad04a87db39c",
		"pipeline-1044864799342689",
		"erda-system", "kube-system", "default",
	}, nil
}

// Mocking the GetAllProjectsDisplayNameFromCache function
func mockGetAllProjectsDisplayNameFromCache(bdl *bundle.Bundle, orgID string) (map[uint64]string, error) {
	return map[uint64]string{
		1: "Project 1",
		2: "Project 2",
	}, nil
}

type mockSteveServer struct {
	cmp.SteveServer
}

func (s *mockSteveServer) ListSteveResource(context.Context, *apistructs.SteveRequest) ([]types.APIObject, error) {
	return []types.APIObject{}, nil
}

func TestSetComponentValue(t *testing.T) {
	filter := &ComponentFilter{
		sdk: &cptype.SDK{
			Lang: make([]*i18n2.LanguageCode, 0),
			Tran: &i18n2.NopTranslator{},
			Identity: &pb.IdentityInfo{
				UserID: "fake-user-id",
				OrgID:  "fake-org-id",
			},
			InParams: map[string]interface{}{
				"clusterName": "local-cluster",
			},
		},
		server: &mockSteveServer{},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, cptype.GlobalInnerKeyCtxSDK, filter.sdk)

	monkey.Patch(cputil.GetAllNamespacesFromCache, mockGetAllNamespacesFromCache)
	monkey.Patch(cputil.GetAllProjectsDisplayNameFromCache, mockGetAllProjectsDisplayNameFromCache)
	defer monkey.UnpatchAll()

	// Call the function to test
	err := filter.SetComponentValue(ctx)

	// Assertions
	assert.NoError(t, err)
}
