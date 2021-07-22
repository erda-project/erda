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

package interface_factory

import (
	"strings"

	"k8s.io/client-go/kubernetes"
)

func IsResourceExist(client *kubernetes.Clientset, resourceKind, groupVersion string) (bool, error) {
	groups, err := client.Discovery().ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return false, err
	}
	for _, resource := range groups.APIResources {
		if strings.EqualFold(resourceKind, resource.Kind) {
			return true, nil
		}
	}
	return false, nil
}
