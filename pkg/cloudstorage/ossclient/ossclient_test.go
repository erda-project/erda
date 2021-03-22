package ossclient

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	BUCKET = "terminus-dice"
)

var c *OssClient

func TestNewClient(t *testing.T) {
	_, err := New("oss-cn-hangzhou.aliyuncs.com", "xxx", "xx")
	assert.Nil(t, err)
}

func TestClient_GetFileUrl(t *testing.T) {
	url, err := c.GetFileUrl(BUCKET, "test.yml")
	assert.Nil(t, err)
	logrus.Info(">>", url)
}

func TestClient_UploadFile(t *testing.T) {
	url, err := c.UploadFile(BUCKET, "yhp-test-alert.sh", "/Users/yuhaiping/Desktop/alert.sh")
	assert.Nil(t, err)
	logrus.Info(">> url:", url)
}

func TestOssClient_DownloadFile(t *testing.T) {
	bs, err := c.DownloadFile(BUCKET, "test.yml")
	assert.Nil(t, err)
	logrus.Info(string(bs))
}
