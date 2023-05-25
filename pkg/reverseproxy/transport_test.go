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

package reverseproxy_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func TestTimerTransport_RoundTrip(t *testing.T) {
	var trans = &reverseproxy.TimerTransport{
		Logger: logrusx.New(logrusx.WithLevel(logrus.DebugLevel)),
		Inner: &reverseproxy.CurlPrinterTransport{
			Logger: logrusx.New(logrusx.WithLevel(logrus.DebugLevel)),
			Inner:  &reverseproxy.DoNothingTransport{},
		},
	}
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080", bytes.NewBufferString("mocked body"))
	for i := 0; i < 10; i++ {
		for j := 0; j < 2; j++ {
			req.Header.Add("key"+strconv.Itoa(i), "value"+strconv.Itoa(i)+strconv.Itoa(j))
		}
	}
	resp, err := trans.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}
