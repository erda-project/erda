// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package docker

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
)

type MockClient struct {
	events chan events.Message
	errs   chan error
}

func (c *MockClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return []types.Container{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
		{ID: "4"},
	}, nil

}

func (c *MockClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if containerID == "1" {
		return types.ContainerJSON{}, errors.New("error")
	} else if containerID == "2" {
		return types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				ID:    containerID,
				Image: "image" + containerID,
			},
			Config: &dockerContainer.Config{
				Env:    []string{"a=123", "b=39-1"},
				Labels: map[string]string{"a": "3241", "c": "8192"},
			},
		}, nil
	}

	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:    containerID,
			Name:  "name" + containerID,
			Image: "image" + containerID,
		},
		Config: &dockerContainer.Config{
			Env:    []string{"a=123", "b=39-1"},
			Labels: map[string]string{"a": "3241", "c": "8192"},
		},
	}, nil
}

func (c *MockClient) Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
	return c.events, c.errs
}

func (c *MockClient) addCreateEvents(id string) {
	c.events <- events.Message{
		Action: "start",
		Actor: events.Actor{
			ID: id,
		},
	}
}

func (c *MockClient) addDieEvents(id string) {
	c.events <- events.Message{
		Action: "die",
		Actor: events.Actor{
			ID: id,
		},
	}
}

func (c *MockClient) addErr(err error) {
	c.errs <- err
}

// func TestWatch(t *testing.T) {
// 	cfg := &Config{
// 		Host:           "unix:///var/run/docker.sock",
// 		RequestTimeout: time.Second,
// 		EventTimeout:   time.Minute,
// 		WatchInterval:  time.Second,
// 		WatchTimeout:   10 * time.Second,
// 		CleanupTimeout: time.Second,
// 	}
// 	c := &MockClient{
// 		events: make(chan events.Message),
// 		errs:   make(chan error),
// 	}
// 	w := newWatchWithClient(cfg, c)
//
// 	w.Start()
// 	defer w.Stop()
//
// 	containers := w.Containers()
// 	if len(containers) != 2 {
// 		t.Fatal("origin containers")
// 	}
//
// 	c.addCreateEvents("1")
// 	containers = w.Containers()
// 	if len(containers) != 2 {
// 		t.Fatal("add error container")
// 	}
//
// 	c.addCreateEvents("5")
// 	time.Sleep(time.Second)
// 	containers = w.Containers()
// 	if len(containers) != 3 {
// 		t.Fatal("add new container")
// 	}
//
// 	c.addDieEvents("5")
// 	time.Sleep(5 * time.Second)
// 	containers = w.Containers()
// 	if len(containers) != 2 {
// 		t.Fatal("delete exist container")
// 	}
//
// 	c.addDieEvents("5")
// 	time.Sleep(5 * time.Second)
// 	containers = w.Containers()
// 	if len(containers) != 2 {
// 		t.Fatal("delete not exist container")
// 	}
// }
