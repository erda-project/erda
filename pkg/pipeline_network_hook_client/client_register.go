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

package pipeline_network_hook_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httputil"
)

var hookClientMap map[string]*apistructs.PipelineLifecycleHookClient

func PostLifecycleHookHttpClient(source string, req interface{}, resp interface{}) error {

	logrus.Debugf("postLifecycleHookHttpClient source: %v, request: %v", source, req)

	if hookClientMap == nil {
		return fmt.Errorf("not find this source: %v client", source)
	}

	client := hookClientMap[source]
	if client == nil {
		return fmt.Errorf("not find this source: %v client", source)
	}

	var httpClient = httpclient.New(
		httpclient.WithTimeout(time.Second, time.Second*5),
	)
	var buffer bytes.Buffer
	r, err := httpClient.Post(client.Host).
		Header(httputil.InternalHeader, "pipeline_lifecycle_hook").
		Path(client.Prefix + "/actions/lifecycle").
		JSONBody(&req).
		Do().
		Body(&buffer)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	if !r.IsOK() {
		return apierrors.ErrInvoke.InternalError(fmt.Errorf("request pipeline lifecycle hook failed httpcode: %v, body: %s", r.StatusCode(), buffer.String()))
	}

	err = json.NewDecoder(&buffer).Decode(resp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(fmt.Errorf("body: %s, decode error %v", buffer.String(), err))
	}

	logrus.Debugf("postLifecycleHookHttpClient response: %v", buffer.String())
	return nil
}

// cache client information
func RegisterLifecycleHookClient(client *dbclient.Client) error {

	list, err := client.FindLifecycleHookClientList()
	if err != nil {
		return fmt.Errorf("not find lifecycleHook hook client list: error %v", err)
	}

	var clientMap = map[string]*apistructs.PipelineLifecycleHookClient{}
	for _, dbHookClient := range list {
		clientMap[dbHookClient.Name] = &apistructs.PipelineLifecycleHookClient{
			ID:     dbHookClient.ID,
			Name:   dbHookClient.Name,
			Host:   dbHookClient.Host,
			Prefix: dbHookClient.Prefix,
		}
	}

	hookClientMap = clientMap
	return nil
}
