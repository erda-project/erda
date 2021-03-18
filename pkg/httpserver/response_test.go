package httpserver

import (
	"net/http"
	"testing"
)

const (
	defaultStatus  = http.StatusAccepted
	defaultContent = "content"
)

func fakeHTTPResponse() Responser {
	return &HTTPResponse{
		Status:  defaultStatus,
		Content: defaultContent,
	}
}

func TestHTTPResponse_GetContent(t *testing.T) {
	responser := fakeHTTPResponse()
	content := responser.GetContent()
	if content != defaultContent {
		t.Errorf("Expected content %s, but got %s", defaultContent, content)
	}
}

func TestHTTPResponse_GetStatus(t *testing.T) {
	responser := fakeHTTPResponse()
	status := responser.GetStatus()
	if status != defaultStatus {
		t.Errorf("Exepcted status code %d, but got %d", defaultStatus, status)
	}
}
