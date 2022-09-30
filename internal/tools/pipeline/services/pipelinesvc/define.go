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

package pipelinesvc

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/clusterinfo"
)

type PipelineSvc struct {
	dbClient *dbclient.Client
	bdl      *bundle.Bundle

	// providers
	clusterInfo clusterinfo.Interface
}

func New(dbClient *dbclient.Client, bdl *bundle.Bundle, clusterInfo clusterinfo.Interface) *PipelineSvc {

	s := PipelineSvc{}
	s.dbClient = dbClient
	s.bdl = bdl
	s.clusterInfo = clusterInfo
	return &s
}
