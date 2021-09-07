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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"time"

	"github.com/pkg/errors"
	"github.com/recallsong/go-utils/container/slice"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-podDetail/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (t *ComponentEventTable) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, globalStateData *cptype.GlobalStateData) error {
	if err := t.SetCtxBundle(ctx); err != nil {
		return fmt.Errorf("failed to set eventTable component ctx bundle, %v", err)
	}
	if err := t.GenComponentState(component); err != nil {
		return err
	}

	// set page no. and page size in first render
	if event.Operation != common.CMPDashboardChangePageNoOperationKey {
		t.State.PageNo = 1
	}
	if event.Operation == cptype.InitializeOperation {
		t.State.PageSize = 20
	}
	// set page no. if triggered by filter
	if err := t.DecodeURLQuery(); err != nil {
		return fmt.Errorf("failed to decode url query for eventTable component, %v", err)
	}
	if err := t.RenderList(); err != nil {
		return err
	}
	if err := t.EncodeURLQuery(); err != nil {
		return fmt.Errorf("failed to encode url query for eventTable component, %v", err)
	}
	t.SetComponentValue()
	return nil
}

func (t *ComponentEventTable) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	if bdl == nil {
		return errors.New("context bundle can not be empty")
	}
	t.ctxBdl = bdl
	t.SDK = cputil.SDK(ctx)
	return nil
}

func (t *ComponentEventTable) GenComponentState(component *cptype.Component) error {
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

func (t *ComponentEventTable) DecodeURLQuery() error {
	queryData, ok := t.SDK.InParams["eventTable__urlQuery"].(string)
	if !ok {
		return nil
	}
	decode, err := base64.StdEncoding.DecodeString(queryData)
	if err != nil {
		return err
	}
	query := make(map[string]int)
	if err := json.Unmarshal(decode, &query); err != nil {
		return err
	}
	t.State.PageNo = uint64(query["pageNo"])
	t.State.PageSize = uint64(query["pageSize"])
	return nil
}

func (t *ComponentEventTable) EncodeURLQuery() error {
	query := make(map[string]int)
	query["pageNo"] = int(t.State.PageNo)
	query["pageSize"] = int(t.State.PageSize)

	data, err := json.Marshal(query)
	if err != nil {
		return err
	}

	encode := base64.StdEncoding.EncodeToString(data)
	t.State.EventTableUQLQuery = encode
	return nil
}

func (t *ComponentEventTable) RenderList() error {
	userID := t.SDK.Identity.UserID
	orgID := t.SDK.Identity.OrgID

	req := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        apistructs.K8SEvent,
		ClusterName: t.State.ClusterName,
	}

	obj, err := t.ctxBdl.ListSteveResource(&req)
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

			Message: fields[6],
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
		Columns: []Column{
			{
				DataIndex: "lastSeen",
				Title:     "Last Seen",
				Width:     160,
				Sorter:    true,
			},
			{
				DataIndex: "type",
				Title:     "Event Type",
				Width:     100,
				Sorter:    true,
			},
			{
				DataIndex: "reason",
				Title:     "Reason",
				Width:     100,
				Sorter:    true,
			},
			{
				DataIndex: "message",
				Title:     "Message",
				Width:     120,
				Sorter:    false,
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
	if l >= length {
		l = 0
	}
	r := l + pageSize
	if r > length {
		r = length
	}
	return l, r
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "eventTable", func() servicehub.Provider {
		return &ComponentEventTable{}
	})
}
