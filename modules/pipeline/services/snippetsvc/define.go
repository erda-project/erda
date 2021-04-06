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

package snippetsvc

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
)

type SnippetSvc struct {
	dbClient *dbclient.Client
	bdl      *bundle.Bundle
}

func New(dbClient *dbclient.Client, bdl *bundle.Bundle) *SnippetSvc {
	s := SnippetSvc{}
	s.dbClient = dbClient
	s.bdl = bdl
	return &s
}
