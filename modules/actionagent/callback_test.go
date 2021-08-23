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

package actionagent

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
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

func TestAppendMetaFile(t *testing.T) {
	var kvs = []struct {
		key   string
		value string
	}{
		{
			key: "name",
			value: `aaa
bbb`,
		},
		{
			key: "name1",
			value: `123456
2345667`,
		},
		{
			key:   "name3",
			value: `aaabbb`,
		},
	}

	var fields []*apistructs.MetadataField
	for _, value := range kvs {
		fields = append(fields, &apistructs.MetadataField{Name: value.key, Value: value.value})
	}

	cb := Callback{}
	cb.AppendMetadataFields(fields)
	for index := range kvs {
		assert.Equal(t, cb.Metadata[index].Name, kvs[index].key)
		assert.Equal(t, cb.Metadata[index].Value, kvs[index].value)
	}
}
