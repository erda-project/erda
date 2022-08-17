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

package collector

import (
	"bytes"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_provider_collectLogsWithSource(t *testing.T) {
	p := &provider{

		Cfg: &config{MetadataKeyOfTopic: "KAFKA-TOPIC"},
	}
	// Setup
	p.consumer = func(data odata.ObservableData) error {
		assert.Equal(t, []byte(`{"source": "job"}`), data.(*odata.Raw).Data)
		assert.Equal(t, map[string]string{"KAFKA-TOPIC": "spot-job-log"}, data.(*odata.Raw).Meta)
		return nil
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(`[{"source": "job"}]`)))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	assert.Nil(t, p.collectLogsWithSource(c, ""))

	// Setup
	p.consumer = func(data odata.ObservableData) error {
		assert.Equal(t, []byte(`{"source": "container"}`), data.(*odata.Raw).Data)
		assert.Equal(t, map[string]string{"KAFKA-TOPIC": "spot-container-log"}, data.(*odata.Raw).Meta)
		return nil
	}
	e = echo.New()
	req = httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(`[{"source": "container"}]`)))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	assert.Nil(t, p.collectLogsWithSource(c, ""))
}
