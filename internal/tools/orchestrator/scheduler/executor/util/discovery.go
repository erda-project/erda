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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
)

func VersionHas(c kubernetes.Interface, v string) (bool, error) {
	serverVersions := sets.String{}
	groups, err := c.Discovery().ServerGroups()
	if err != nil {
		return false, err
	}

	versions := metav1.ExtractGroupVersions(groups)
	for _, v := range versions {
		serverVersions.Insert(v)
	}

	return serverVersions.Has(v), nil
}
