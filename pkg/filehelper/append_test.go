package filehelper

import (
	"io/ioutil"
	"os"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestAppend(t *testing.T) {
	tmpDir := os.TempDir()
	f, err := ioutil.TempFile(tmpDir, "")
	assert.NoError(t, err)
	defer func() {
		f.Close()
		os.RemoveAll(tmpDir)
	}()
	assert.NoError(t, Append(f.Name(), "action meta: hello=world\n"))
	assert.NoError(t, Append(f.Name(), "action meta: key=value\n"))
	b, err := ioutil.ReadAll(f)
	assert.NoError(t, err)
	t.Log("write: ", string(b))
}
