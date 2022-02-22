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

package kafka

import (
	"encoding/json"
	"testing"
)

func TestXXX(t *testing.T) {
	data := []byte("{\"date\":\"2018-05-30T09:39:52.000681Z\",\"log\":\"hello world\\n\",\"stream\":\"stdout\",\"time\":\"2021-08-16T07:56:01.025279973Z\",\"level\":\"INFO\",\"source\":\"container\",\"kubernetes\":{\"pod_name\":\"pod\",\"namespace\":\"ns\",\"container_name\":\"cname\",\"container_id\":\"cid\",\"pod_id\":\"pid\",\"labels\":{\"k\":\"v\"},\"annotations\":{\"k\":\"v\"}}}")
	buf, _ := json.Marshal(data)
	t.Log(string(buf))
}
