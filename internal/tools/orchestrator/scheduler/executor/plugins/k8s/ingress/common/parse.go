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

package common

import "strings"

const (
	LabelHAProxyVHost = "HAPROXY_0_VHOST"
)

func ParsePublicHostsFromLabel(labels map[string]string) []string {
	v, ok := labels[LabelHAProxyVHost]
	// No external domain name
	if !ok {
		return []string{}
	}
	// Forward the domain name/vip set corresponding to HAPROXY_0_VHOST in the label to the 0th port of the service
	return strings.Split(v, ",")
}
