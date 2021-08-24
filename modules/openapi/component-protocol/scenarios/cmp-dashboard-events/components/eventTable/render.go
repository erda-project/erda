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

package eventTable

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/recallsong/go-utils/container/slice"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func RenderCreator() protocol.CompRender {
	return &ComponentEventTable{}
}

func (t *ComponentEventTable) Render(ctx context.Context, component *apistructs.Component, _ apistructs.ComponentProtocolScenario,
	event apistructs.ComponentEvent, globalStateData *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil {
		return errors.New("context bundle can not be empty")
	}
	t.ctxBdl = bdl
	if err := t.GenComponentState(component); err != nil {
		return err
	}

	// set page no. and page size in first render
	if event.Operation == apistructs.InitializeOperation {
		t.State.PageNo = 1
		t.State.PageSize = 20
	}
	// set page no. if triggered by filter
	if event.Operation == apistructs.RenderingOperation || event.Operation == apistructs.OnChangeSortOperation ||
		event.Operation == apistructs.ChangeOrgsPageSizeOperationKey {
		t.State.PageNo = 1
	}
	if err := t.RenderList(); err != nil {
		return err
	}
	t.SetComponentValue()
	return nil
}

func (t *ComponentEventTable) GenComponentState(component *apistructs.Component) error {
	if component == nil || component.State == nil {
		return nil
	}

	data, err := json.Marshal(component.State)
	if err != nil {
		logrus.Errorf("failed to marshal for eventTable state, %v", err)
		return err
	}
	var state State
	err = json.Unmarshal(data, &state)
	if err != nil {
		logrus.Errorf("failed to unmarshal for eventTable state, %v", err)
		return err
	}
	t.State = state
	return nil
}

func (t *ComponentEventTable) RenderList() error {
	userID := t.ctxBdl.Identity.UserID
	orgID := t.ctxBdl.Identity.OrgID

	req := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        apistructs.K8SEvent,
		ClusterName: t.State.ClusterName,
	}

	obj, err := t.ctxBdl.Bdl.ListSteveResource(&req)
	if err != nil {
		return err
	}
	list := obj.Slice("data")

	var items []Item
	for _, obj := range list {
		if t.State.FilterValues.Namespace != nil && !contain(t.State.FilterValues.Namespace, obj.String("metadata", "namespace")) {
			continue
		}
		if t.State.FilterValues.Type != nil && !contain(t.State.FilterValues.Type, obj.String("_type")) {
			continue
		}
		fields := obj.StringSlice("metadata", "fields")
		if len(fields) != 10 {
			logrus.Errorf("length of event fields is invalid: %d", len(fields))
			continue
		}
		count, err := strconv.ParseInt(fields[8], 10, 64)
		if err != nil {
			logrus.Errorf("failed to parse count for event %s, %v", fields[9], err)
			continue
		}
		lastSeenTimestamp, err := time.ParseDuration(fields[0])
		if err != nil {
			logrus.Errorf("failed to parse timestamp for event %s, %v", fields[9], err)
			continue
		}
		items = append(items, Item{
			LastSeen:          fields[0],
			LastSeenTimestamp: lastSeenTimestamp.Nanoseconds(),
			Type:              fields[1],
			Reason:            fields[2],
			Object:            fields[3],
			Source:            fields[5],
			Message:           fields[6],
			Count:             fields[8],
			CountNum:          count,
			Name:              fields[9],
			Namespace:         obj.String("metadata", "namespace"),
		})
	}
	if t.State.Sorter.Field != "" {
		cmpWrapper := func(field, order string) func(int, int) bool {
			ascend := order == "ascend"
			switch field {
			case "lastSeen":
				return func(i int, j int) bool {
					less := items[i].LastSeenTimestamp < items[j].LastSeenTimestamp
					if ascend {
						return less
					}
					return !less
				}
			case "type":
				return func(i int, j int) bool {
					less := items[i].Type < items[j].Type
					if ascend {
						return less
					}
					return !less
				}
			case "reason":
				return func(i int, j int) bool {
					less := items[i].Reason < items[j].Reason
					if ascend {
						return less
					}
					return !less
				}
			case "object":
				return func(i int, j int) bool {
					less := items[i].Object < items[j].Object
					if ascend {
						return less
					}
					return !less
				}
			case "source":
				return func(i int, j int) bool {
					less := items[i].Source < items[j].Source
					if ascend {
						return less
					}
					return !less
				}
			case "message":
				return func(i int, j int) bool {
					less := items[i].Message < items[j].Message
					if ascend {
						return less
					}
					return !less
				}
			case "count":
				return func(i int, j int) bool {
					less := items[i].CountNum < items[j].CountNum
					if ascend {
						return less
					}
					return !less
				}
			case "name":
				return func(i int, j int) bool {
					less := items[i].Name < items[j].Name
					if ascend {
						return less
					}
					return !less
				}
			case "namespace":
				return func(i int, j int) bool {
					less := items[i].Namespace < items[j].Namespace
					if ascend {
						return less
					}
					return !less
				}
			default:
				return func(i int, j int) bool {
					return false
				}
			}
		}
		slice.Sort(items, cmpWrapper(t.State.Sorter.Field, t.State.Sorter.Order))
	}

	l, r := getRange(len(items), int(t.State.PageNo), int(t.State.PageSize))
	t.Data.List = items[l:r]
	t.State.Total = uint64(len(items))
	return nil
}

func (t *ComponentEventTable) SetComponentValue() {
	t.Props = Props{
		PageSizeOptions: []string{"10", "20", "50", "100"},
		Columns:      []Column{
			{
				DataIndex: "lastSeen",
				Title:     "Last Seen",
				Width:     "160",
				Sorter:    true,
			},
			{
				DataIndex: "type",
				Title:     "Event Type",
				Width:     "100",
				Sorter:    true,
			},
			{
				DataIndex: "reason",
				Title:     "Reason",
				Width:     "100",
				Sorter:    true,
			},
			{
				DataIndex: "object",
				Title:     "Object",
				Width:     "150",
				Sorter:    true,
			},
			{
				DataIndex: "source",
				Title:     "Source",
				Width:     "120",
				Sorter:    true,
			},
			{
				DataIndex: "message",
				Title:     "Message",
				Width:     "120",
				Sorter:    true,
			},
			{
				DataIndex: "count",
				Title:     "Count",
				Width:     "80",
				Sorter:    true,
			},
			{
				DataIndex: "name",
				Title:     "Name",
				Width:     "120",
				Sorter:    true,
			},
			{
				DataIndex: "namespace",
				Title:     "Namespace",
				Width:     "120",
				Sorter:    true,
			},
		},   
	}
	t.Operations = make(map[string]interface{})
	t.Operations[apistructs.OnChangePageNoOperation.String()] = Operation{
		Key:    apistructs.OnChangePageNoOperation.String(),
		Reload: true,
	}
	t.Operations[apistructs.OnChangePageSizeOperation.String()] = Operation{
		Key:    apistructs.OnChangePageSizeOperation.String(),
		Reload: true,
	}
	t.Operations[apistructs.OnChangeSortOperation.String()] = Operation{
		Key:    apistructs.OnChangeSortOperation.String(),
		Reload: true,
	}
}

func contain(arr []string, target string) bool {
	for _, str := range arr {
		if target == str {
			return true
		}
	}
	return false
}

func getRange(length, pageNo, pageSize int) (int, int) {
	l := (pageNo - 1) * pageSize
	if l > length {
		l = 0
	}
	r := l + pageSize
	if r > length {
		r = length
	}
	return l, r
}
