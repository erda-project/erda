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

package actionagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleMetaFile(t *testing.T) {
	cb := Callback{}
	err := cb.HandleMetaFile([]byte(`{"metadata":[{"name":"commit","value":"7fbca4b1d73c2a68cf0817c31a0c246747e5735d"},{"name":"author","value":"linjun"},{"name":"author_date","value":"2019-08-07 09:57:52 +0800","type":"time"},{"name":"branch","value":"release/3.4"},{"name":"message","value":"update .npmrc\n","type":"message"}]}{"metadata":[{"name":"commit","value":"7fbca4b1d73c2a68cf0817c31a0c246747e5735d"},{"name":"author","value":"linjun"},{"name":"author_date","value":"2019-08-07 09:57:52 +0800","type":"time"},{"name":"branch","value":"release/3.4"},{"name":"message","value":"update .npmrc\n","type":"message"}]}`))
	assert.NoError(t, err)

	cb = Callback{}
	err = cb.HandleMetaFile([]byte(`
commit= 777
author =xxx
message = test metafile
name=pipeline
`))
	assert.NoError(t, err)
	assert.Equal(t, 4, len(cb.Metadata))
	assert.Equal(t, "commit", cb.Metadata[0].Name)
	assert.Equal(t, "777", cb.Metadata[0].Value)
	assert.Equal(t, "author", cb.Metadata[1].Name)
	assert.Equal(t, "xxx", cb.Metadata[1].Value)
	assert.Equal(t, "message", cb.Metadata[2].Name)
	assert.Equal(t, "test metafile", cb.Metadata[2].Value)
	assert.Equal(t, "name", cb.Metadata[3].Name)
	assert.Equal(t, "pipeline", cb.Metadata[3].Value)
}
