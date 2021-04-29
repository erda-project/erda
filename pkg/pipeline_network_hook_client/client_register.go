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
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httputil"
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

// cache client information in the database every 2 hours
func RegisterLifecycleHookClient(client *dbclient.Client) {

	list, err := client.FindLifecycleHookClientList()
	if err != nil {
		logrus.Errorf("not find lifecycleHook hook client list: error %v", err)
		return
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
}
