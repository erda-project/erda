package cache

import (
	"context"
	"github.com/sirupsen/logrus"
)

type EventMethod string

const EventMethodSet EventMethod = "Set"
const EventMethodRemove EventMethod = "Remove"
const EventMethodClear EventMethod = "Clear"

type Event struct {
	Method EventMethod
	Name   string
	Tag    string
	Data   *McpServerInfo
}

type EventListener struct {
	events chan *Event
}

func NewEventListener() *EventListener {
	return &EventListener{
		events: make(chan *Event),
	}
}

func (e *EventListener) Send(ctx context.Context, event *Event) {
	go func(ctx context.Context) {
		select {
		case e.events <- event:
			return
		case <-ctx.Done():
			logrus.Errorf("Send event context canceled")
			return
		}
	}(ctx)
}

func (e *EventListener) OnEvent(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			select {
			case event := <-e.events:
				switch event.Method {
				case EventMethodSet:
					if err := SetMcpServer(event.Name, event.Tag, event.Data); err != nil {
						logrus.Errorf("OnEvent SetMcpServer error: %v", err)
					}
				case EventMethodRemove:
					if err := RemoveMcpServer(event.Name, event.Tag); err != nil {
						logrus.Errorf("OnEvent RemoveMcpServer error: %v", err)
					}
				case EventMethodClear:
					if err := ClearMcpServers(); err != nil {
						logrus.Errorf("OnEvent ClearMcpServers error: %v", err)
					}
				}
			case <-ctx.Done():
				logrus.Errorf("Send event context canceled")
				return
			}
		}
	}(ctx)
}
