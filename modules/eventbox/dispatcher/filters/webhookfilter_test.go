package filters

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/erda-project/erda/modules/eventbox/constant"
	"github.com/erda-project/erda/modules/eventbox/types"
	"github.com/erda-project/erda/modules/eventbox/webhook"

	"github.com/stretchr/testify/assert"
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
