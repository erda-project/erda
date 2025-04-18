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

package util

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda/apistructs"
)

func TestParsePreserveProjects(t *testing.T) {
	key := "PRESERVEPROJECTS"

	assert.Equal(t, map[string]struct{}{
		"1": {}, "2": {}, "3": {},
	}, ParsePreserveProjects(map[string]string{
		key: "1,2,3",
	}, key))

	assert.Equal(t, map[string]struct{}{}, ParsePreserveProjects(nil, key))
}

func TestBuildDcosConstraints(t *testing.T) {
	assert.Equal(t, [][]string{}, BuildDcosConstraints(false, nil, nil, nil))
	assert.Equal(t, [][]string{}, BuildDcosConstraints(false, map[string]string{
		"MATCH_TAGS":   "",
		"EXCLUDE_TAGS": "locked,platform",
	}, nil, nil))
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bany\b.*`},
	}, BuildDcosConstraints(true, nil, nil, nil))
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bany\b.*`},
	}, BuildDcosConstraints(true, map[string]string{}, nil, nil))
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bany\b.*`},
		{"dice_tags", "UNLIKE", `.*\blocked\b.*`},
		{"dice_tags", "UNLIKE", `.*\bplatform\b.*`},
	}, BuildDcosConstraints(true, map[string]string{
		"MATCH_TAGS":   "",
		"EXCLUDE_TAGS": "locked,platform",
	}, nil, nil))
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bany\b.*`},
		{"dice_tags", "UNLIKE", `.*\block1\b.*`},
	}, BuildDcosConstraints(true, map[string]string{
		"EXCLUDE_TAGS": "lock1",
	}, nil, nil))
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bpack1\b.*`},
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bpack2\b.*`},
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bpack3\b.*`},
		{"dice_tags", "UNLIKE", `.*\block1\b.*`},
		{"dice_tags", "UNLIKE", `.*\block2\b.*`},
		{"dice_tags", "UNLIKE", `.*\block3\b.*`},
	}, BuildDcosConstraints(true, map[string]string{
		"MATCH_TAGS":   "pack1,pack2,pack3",
		"EXCLUDE_TAGS": "lock1,lock2,lock3",
	}, nil, nil))

	// test preserve project
	assert.Equal(t, [][]string{
		{"dice_tags", "LIKE", `.*\bproject-32\b.*`},
		{"dice_tags", "LIKE", `.*\bt1\b.*`},
		{"dice_tags", "LIKE", `.*\bt2\b.*`},
		{"dice_tags", "LIKE", `.*\bt3\b.*`},
		{"dice_tags", "UNLIKE", `.*\be1\b.*`},
		{"dice_tags", "UNLIKE", `.*\be2\b.*`},
		{"dice_tags", "UNLIKE", `.*\be3\b.*`},
	}, BuildDcosConstraints(true, map[string]string{
		"MATCH_TAGS":   "t1,t2,t3",
		"EXCLUDE_TAGS": "e1,e2,e3",
		"DICE_PROJECT": "32",
	}, map[string]struct{}{
		"32": {},
	}, nil))

	// test not preserve project
	assert.Equal(t, [][]string{
		{"dice_tags", "UNLIKE", `.*\bproject-[^,]+\b.*`},
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bt1\b.*`},
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bt2\b.*`},
		{"dice_tags", "LIKE", `.*\bany\b.*|.*\bt3\b.*`},
		{"dice_tags", "UNLIKE", `.*\be1\b.*`},
		{"dice_tags", "UNLIKE", `.*\be2\b.*`},
		{"dice_tags", "UNLIKE", `.*\be3\b.*`},
	}, BuildDcosConstraints(true, map[string]string{
		"MATCH_TAGS":   "t1,t2,t3",
		"EXCLUDE_TAGS": "e1,e2,e3",
		"DICE_PROJECT": "32",
	}, map[string]struct{}{}, nil))
}

func TestCombineLabels(t *testing.T) {
	assert.Equal(t, map[string]string{}, CombineLabels(nil, nil))

	assert.Equal(t, map[string]string{
		"A": "v1",
		"B": "v4",
		"C": "v3",
	}, CombineLabels(map[string]string{
		"A": "v1",
		"B": "v2",
	}, map[string]string{
		"C": "v3",
		"B": "v4",
	}))
}

func TestGetClient(t *testing.T) {
	clusterName := "fake-clusterName"
	fakeAddress := "fake-address"
	fakeToken := "fake-token"

	_, _, err := GetClient(clusterName, nil)
	assert.Equal(t, err, fmt.Errorf("cluster %s manage config is nil", clusterName))

	_, _, err = GetClient(clusterName, &apistructs.ManageConfig{
		Type:    apistructs.ManageProxy,
		Address: fakeAddress,
	})
	assert.Equal(t, err, fmt.Errorf("token or address is empty"))

	mc1 := &apistructs.ManageConfig{
		Type:    apistructs.ManageProxy,
		Address: fakeAddress,
		Token:   fakeToken,
	}
	address, _, err := GetClient(clusterName, mc1)
	assert.Equal(t, strings.Contains(address, "inet://"), true)
	assert.Equal(t, err, nil)

	mc1.Type = apistructs.ManageToken
	address, _, err = GetClient(clusterName, mc1)
	assert.Equal(t, strings.Contains(address, "inet://"), false)
	assert.Equal(t, err, nil)

	_, _, err = GetClient(clusterName, &apistructs.ManageConfig{
		Type: apistructs.ManageCert,
	})
	assert.Equal(t, err, fmt.Errorf("cert or key is empty"))

	_, _, err = GetClient(clusterName, &apistructs.ManageConfig{})
	assert.Equal(t, err, fmt.Errorf("manage type is not support"))

	m2 := &apistructs.ManageConfig{
		Type: apistructs.ManageCert,
	}

	_, _, err = GetClient(clusterName, m2)
	assert.Equal(t, err, fmt.Errorf("cert or key is empty"))
}

func Test_ParseAnnotationFromEnv(t *testing.T) {
	type args struct {
		name string
		key  string
		want string
	}

	tests := []args{
		{
			name: "dice component env",
			key:  "DICE_ORG_ID",
			want: "msp.erda.cloud/org_id",
		},
		{
			name: "dice addon env",
			key:  "N0_DICE_ORG_ID",
			want: "msp.erda.cloud/org_id",
		},
		{
			name: "msp env",
			key:  "MSP_LOG_ATTACH",
			want: "msp.erda.cloud/msp_log_attach",
		},
		{
			name: "other key",
			key:  "TERMINUS_KEY",
			want: "msp.erda.cloud/terminus_key",
		},
		{
			name: "empty key",
			key:  "HELLO_WORLD",
			want: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ParseAnnotationFromEnv(test.key)
			assert.Equal(t, got, test.want)
		})
	}
}

func TestResourceFormatters(t *testing.T) {
	testCases := []struct {
		name      string
		value     interface{}
		formatter func(interface{}) resource.Quantity
		expected  string
	}{
		{
			name:  "CPU Formatter with int",
			value: 1000,
			formatter: func(v interface{}) resource.Quantity {
				return ResourceCPUFormatter(v.(int))
			},
			expected: "1",
		},
		{
			name:  "CPU Formatter with float",
			value: 1000.5,
			formatter: func(v interface{}) resource.Quantity {
				return ResourceCPUFormatter(v.(float64))
			},
			expected: "1000500u",
		},
		{
			name:  "Memory Formatter with int",
			value: 512,
			formatter: func(v interface{}) resource.Quantity {
				return ResourceMemoryFormatter(v.(int))
			},
			expected: "512Mi",
		},
		{
			name:  "Memory Formatter with float",
			value: 512.5,
			formatter: func(v interface{}) resource.Quantity {
				return ResourceMemoryFormatter(v.(float64))
			},
			expected: "524800Ki",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.formatter(tc.value)
			if result.String() != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result.String())
			}
		})
	}
}
