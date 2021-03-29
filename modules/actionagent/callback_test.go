package actionagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleMetaFile(t *testing.T) {
	cb := callback{}
	err := cb.handleMetaFile([]byte(`{"metadata":[{"name":"commit","value":"7fbca4b1d73c2a68cf0817c31a0c246747e5735d"},{"name":"author","value":"linjun"},{"name":"author_date","value":"2019-08-07 09:57:52 +0800","type":"time"},{"name":"branch","value":"release/3.4"},{"name":"message","value":"update .npmrc\n","type":"message"}]}{"metadata":[{"name":"commit","value":"7fbca4b1d73c2a68cf0817c31a0c246747e5735d"},{"name":"author","value":"linjun"},{"name":"author_date","value":"2019-08-07 09:57:52 +0800","type":"time"},{"name":"branch","value":"release/3.4"},{"name":"message","value":"update .npmrc\n","type":"message"}]}`))
	assert.NoError(t, err)

	cb = callback{}
	err = cb.handleMetaFile([]byte(`
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
