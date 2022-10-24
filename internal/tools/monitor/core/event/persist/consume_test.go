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

package persist

import (
	"testing"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/tools/monitor/core/event"
)

var data = `
{
    "eventID": "a7cdf68b-ab27-4ee2-8dc8-4ffcacbcxxxx",
    "severity": null,
    "name": "exception",
    "kind": "EVENT_KIND_EXCEPTION",
    "timeUnixNano": 1666319382526000000,
    "relations":
    {
        "traceID": null,
        "resID": "857282ecaf412871d06cdc0c411xxxx",
        "resType": "exception",
        "resourceKeys": null
    },
    "attributes":
    {
        "metaData": "{\"message\":\"Exception thrown by subscriber method updateUserSendMqSubscribe(io.terminus.draco.server.event.issue.UpdateUserSendMqIssue) on subscriber io.terminus.draco.server.event.subscribe.UserSendMqIssueSubscriber@6f7be76 when dispatching event: UpdateUserSendMqIssue(userId=12345678)\",\"type\":\"java.lang.NullPointerException\"}",
        "requestHeaders": "{}",
        "requestContext": "{}",
        "requestId": null,
        "terminusKey": "n77506fb33bda44fdb094b030539b9029",
        "tags": "{\"cluster_name\":\"test\",\"host_ip\":\"0.0.0.1\",\"workspace\":\"PROD\",\"service_name\":\"server\",\"runtime_id\":\"123456\",\"language\":\"Java\",\"service_instance_id\":\"5034b444-fb5e-45f7-822c-f60ba95bxxxx\",\"project_name\":\"test-project\",\"application_id\":\"1234\",\"terminus_key\":\"n77506fb33bda44fdb094b030539bxxxx\",\"runtime_name\":\"master\",\"event_id\":\"a7cdf68b-ab27-4ee2-8dc8-4ffcacbcxxxx\",\"application_name\":\"test-app\",\"instance_id\":\"5034b444-fb5e-45f7-822c-f60ba95bxxxx\",\"project_id\":\"123\",\"service_id\":\"1234\",\"host\":\"node-xxx\",\"org_name\":\"test-org\"}"
    },
    "message": "[]"
}`

type MockStatistics struct {
	Statistics
}

func (s *MockStatistics) DecodeError(value []byte, err error) {

}

func (s *MockStatistics) ValidateError(data *event.Event) {

}

func (s *MockStatistics) MetadataError(data *event.Event, err error) {

}

type MockLogger struct {
	logs.Logger
	t *testing.T
}

func (l *MockLogger) Warnf(template string, args ...interface{}) {
	l.t.Errorf(template, args...)
}

func (l *MockLogger) Errorf(template string, args ...interface{}) {
	l.t.Errorf(template, args...)
}

type MockValidator struct {
}

func (v *MockValidator) Validate(l *event.Event) error {
	return nil
}

type MockMetadataProcessor struct {
}

func (m *MockMetadataProcessor) Process(data *event.Event) error {
	return nil
}

func TestProvider_Decode(t *testing.T) {
	p := &provider{
		Cfg:       &config{},
		Log:       &MockLogger{t: t},
		stats:     &MockStatistics{},
		validator: &MockValidator{},
		metadata:  &MockMetadataProcessor{},
	}
	_, err := p.decode(nil, []byte(data), nil, time.Now())
	if err != nil {
		t.Error(err)
	}
}
