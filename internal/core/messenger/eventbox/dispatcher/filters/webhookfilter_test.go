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

package filters

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/core/messenger/eventbox/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/messenger/eventbox/constant"
	"github.com/erda-project/erda/internal/core/messenger/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/internal/core/messenger/eventbox/types"
	"github.com/erda-project/erda/internal/core/messenger/eventbox/webhook"
)

func testWebhookFilter(t *testing.T, f Filter) {
	m := types.Message{
		Sender:  "self",
		Content: "233",
		Labels: map[types.LabelKey]interface{}{
			types.LabelKey(constant.WebhookLabelKey): webhook.EventLabel{
				Event:         "test-event",
				OrgID:         "1",
				ProjectID:     "2",
				ApplicationID: "3",
			},
		},
		Time: 0,
	}
	derr := f.Filter(&m)
	assert.True(t, derr.IsOK())

	http := m.Labels[types.LabelKey("/HTTP")]
	assert.NotNil(t, http, fmt.Sprintf("%+v", m))

	raw, err := json.Marshal(m.Content)
	assert.Nil(t, err)
	em := webhook.EventMessage{}
	assert.Nil(t, json.Unmarshal(raw, &em))
	assert.Equal(t, "test-event", em.Event)
	assert.Equal(t, "", em.Env)
}

func testWebhookFilterWithEnv(t *testing.T, f Filter) {
	m := types.Message{
		Sender:  "self",
		Content: "233",
		Labels: map[types.LabelKey]interface{}{
			types.LabelKey(constant.WebhookLabelKey): webhook.EventLabel{
				Event:         "test-event",
				OrgID:         "1",
				ProjectID:     "2",
				ApplicationID: "3",
				Env:           "test",
			},
		},
		Time: 0,
	}
	derr := f.Filter(&m)
	assert.True(t, derr.IsOK())

	http := m.Labels[types.LabelKey("/HTTP")]
	assert.NotNil(t, http, fmt.Sprintf("%+v", m))

	raw, err := json.Marshal(m.Content)
	assert.Nil(t, err)
	em := webhook.EventMessage{}
	assert.Nil(t, json.Unmarshal(raw, &em))
	assert.Equal(t, "test", em.Env)
}

func testWebhookFilterDINGDINGURL(t *testing.T, f Filter) {
	m := types.Message{
		Sender:  "self",
		Content: "2333",
		Labels: map[types.LabelKey]interface{}{
			types.LabelKey(constant.WebhookLabelKey): webhook.EventLabel{
				Event:         "test-event3",
				OrgID:         "1",
				ProjectID:     "2",
				ApplicationID: "3",
				Env:           "test",
			},
		},
	}
	derr := f.Filter(&m)
	assert.True(t, derr.IsOK())

	dd := m.Labels[types.LabelKey("/DINGDING")]
	assert.NotNil(t, dd, fmt.Sprintf("%+v", m))

}

func TestWebhookFilter(t *testing.T) {
	var (
		whf = &WebhookFilter{}
		whi = &webhook.WebHookImpl{}
	)

	hook := &pb.Hook{
		Id:            "1",
		Name:          "test-event",
		Events:        []string{"event1", "event2"},
		Url:           "https://localhost:8080/hook",
		Active:        true,
		OrgID:         "-1",
		ProjectID:     "-1",
		ApplicationID: "-1",
	}

	monkey.PatchInstanceMethod(reflect.TypeOf(whi), "SearchHooks", func(whi *webhook.WebHookImpl, location webhook.HookLocation, event string) []*pb.Hook {
		return []*pb.Hook{hook}
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(whi), "ListHooks", func(whi *webhook.WebHookImpl, location webhook.HookLocation) ([]*pb.Hook, error) {
		return []*pb.Hook{hook}, nil
	})

	defer monkey.UnpatchAll()

	dispatchErr := whf.Filter(&types.Message{
		Sender:  "unit-test",
		Content: "hello webhook",
		Labels: map[types.LabelKey]interface{}{
			types.LabelKey(constant.WebhookLabelKey): &apistructs.EventHeader{
				Event:         "event1",
				Action:        "update",
				OrgID:         "1",
				ProjectID:     "1",
				ApplicationID: "1",
				TimeStamp:     time.Now().String(),
			},
		},
	})

	assert.Equal(t, dispatchErr, errors.New())
}

// func TestWebhookFilter(t *testing.T) {
// 	impl, err := webhook.NewWebHookImpl()
// 	assert.Nil(t, err)
// 	r, err := impl.CreateHook("1", webhook.CreateHookRequest{
// 		Name:   "xxx",
// 		Events: []string{"test-event", "test-event2"},
// 		URL:    "http://test-url",
// 		Active: true,
// 		HookLocation: apistructs.HookLocation{
// 			Org:         "1",
// 			Project:     "2",
// 			Application: "3",
// 			Env:         []string{"test"},
// 		},
// 	})
// 	assert.Nil(t, err)
// 	r2, err := impl.CreateHook("1", webhook.CreateHookRequest{
// 		Name:   "yyy",
// 		Events: []string{"test-event3", "test-event4"},
// 		URL:    "https://oapi.dingtalk.com/robot/send?access_token=xxxx",
// 		Active: true,
// 		HookLocation: apistructs.HookLocation{
// 			Org:         "1",
// 			Project:     "2",
// 			Application: "3",
// 			Env:         []string{"test"},
// 		},
// 	})
// 	assert.Nil(t, err)
// 	defer func() {
// 		assert.Nil(t, impl.DeleteHook("1", string(r)))
// 		assert.Nil(t, impl.DeleteHook("1", string(r2)), string(r2))
// 	}()
// 	f, err := NewWebhookFilter()
// 	assert.Nil(t, err)

// 	t.Run("normal", func(t *testing.T) {
// 		testWebhookFilter(t, f)
// 	})
// 	t.Run("env", func(t *testing.T) {
// 		testWebhookFilterWithEnv(t, f)
// 	})
// 	t.Run("dingding", func(t *testing.T) {
// 		testWebhookFilterDINGDINGURL(t, f)
// 	})
// }
