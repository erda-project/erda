package mcp_server_director

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestMessage(t *testing.T) {
	id, err := parseSseMessagePath("/message?sessionId=5c4d91f4-ad02-49da-8ec7-b7a43ffce951")
	logrus.Infof("id: %v, err: %v", id, err)
}
