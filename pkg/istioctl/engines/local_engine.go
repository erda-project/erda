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

package engines

import (
	"github.com/erda-project/erda/pkg/clientgo"
	"github.com/erda-project/erda/pkg/istioctl"
	"github.com/erda-project/erda/pkg/istioctl/executors"
)

type LocalEngine struct {
	istioctl.DefaultEngine
}

func NewLocalEngine(addr string) (*LocalEngine, error) {
	client, err := clientgo.New(addr)
	if err != nil {
		return nil, err
	}
	authN := &executors.AuthNExecutor{}
	authN.SetIstioClient(client.CustomClient)
	return &LocalEngine{
		DefaultEngine: istioctl.NewDefaultEngine(authN),
	}, nil
}
