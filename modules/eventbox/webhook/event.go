package webhook

import (
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
)

// message which response to hooked url
type EventMessage struct {
	Event         string `json:"event"`
	Action        string `json:"action"`
	OrgID         string `json:"orgID"`
	ProjectID     string `json:"projectID"`
	ApplicationID string `json:"applicationID"`
	Env           string `json:"env"`
	// content 结构跟具体 event 相关
	Content   json.RawMessage `json:"content"`
	TimeStamp string          `json:"timestamp"`
}

func PingEvent(org, project, application string, h Hook) (*EventMessage, error) {
	hraw, err := json.Marshal(h)
	if err != nil {
		return nil, err
	}
	m := MkEventMessage(EventLabel{
		Event:         "ping",
		Action:        "ping",
		OrgID:         org,
		ProjectID:     project,
		ApplicationID: application},
		hraw)
	return &m, nil
}

func MkEventMessage(label EventLabel, content []byte) EventMessage {
	return EventMessage{
		Event:         label.Event,
		Action:        label.Action,
		OrgID:         label.OrgID,
		ProjectID:     label.ProjectID,
		ApplicationID: label.ApplicationID,
		Env:           label.Env,
		Content:       content,
		TimeStamp:     nowTimestamp(),
	}
}

type EventLabel = apistructs.EventHeader
