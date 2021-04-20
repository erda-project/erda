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

package pipelinesvc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/httpclient"
)

func TestPipelineSvc_tryGetUser(t *testing.T) {
	s := &PipelineSvc{bdl: bundle.New(bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second, time.Second))))}
	invalidUserID := "invalid user id"
	user := s.tryGetUser(invalidUserID)
	assert.Equal(t, invalidUserID, user.ID)
	assert.Empty(t, user.Name)
}
