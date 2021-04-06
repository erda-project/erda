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
