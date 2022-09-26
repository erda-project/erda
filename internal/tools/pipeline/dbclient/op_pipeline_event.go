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

package dbclient

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
)

// GetPipelineEvents get pipeline events from reports.
// return: report, events, error
func (client *Client) GetPipelineEvents(pipelineID uint64, ops ...SessionOption) (*spec.PipelineReport, []*basepb.PipelineEvent, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var report spec.PipelineReport
	report.PipelineID = pipelineID
	report.Type = apistructs.PipelineReportTypeEvent
	exist, err := session.Get(&report)
	if err != nil {
		return nil, nil, err
	}
	if !exist {
		return nil, nil, nil
	}
	if report.Meta == nil {
		return &report, nil, nil
	}
	eventsI, ok := report.Meta[apistructs.PipelineReportEventMetaKey]
	if !ok {
		return &report, nil, nil
	}
	b, err := json.Marshal(eventsI)
	if err != nil {
		return &report, nil, nil
	}
	var events []*basepb.PipelineEvent
	if err := json.Unmarshal(b, &events); err != nil {
		return &report, nil, nil
	}
	return &report, events, nil
}

func (client *Client) AppendPipelineEvent(pipelineID uint64, newEvents []*basepb.PipelineEvent, ops ...SessionOption) error {
	if len(newEvents) == 0 {
		return nil
	}

	session := client.NewSession(ops...)
	defer session.Close()

	// get events
	report, events, err := client.GetPipelineEvents(pipelineID, ops...)
	if err != nil {
		return err
	}

	// get order events
	ordered := makeOrderEvents(events, newEvents)

	if len(ordered) <= 0 {
		return nil
	}

	if report == nil {
		// create
		return client.CreatePipelineReport(&spec.PipelineReport{
			PipelineID: pipelineID,
			Type:       apistructs.PipelineReportTypeEvent,
			Meta:       makeEventReportMeta(ordered),
			CreatorID:  ordered[0].Source.Component,
			UpdaterID:  ordered[0].Source.Component,
		}, ops...)
	}
	// update
	report.Meta = makeEventReportMeta(ordered)
	return client.UpdatePipelineReport(report, ops...)
}

func getLastEvent(ordered orderedEvents) *basepb.PipelineEvent {
	if ordered != nil {
		sort.Sort(ordered)
		return ordered[len(ordered)-1]
	}
	return nil
}

type orderedEvents []*basepb.PipelineEvent

func makeOrderEvents(events []*basepb.PipelineEvent, newEvents []*basepb.PipelineEvent) orderedEvents {
	// order events
	var ordered orderedEvents
	for _, g := range events {
		ordered = append(ordered, g)
	}

	// new events order, if event not have time to add time
	var newEventOrder orderedEvents
	now := time.Now()
	for index, g := range newEvents {
		if g.FirstTimestamp == nil || g.FirstTimestamp.AsTime().IsZero() {
			g.FirstTimestamp = timestamppb.New(now.Add(time.Duration(index) * time.Millisecond))
		}
		if g.LastTimestamp == nil || g.LastTimestamp.AsTime().IsZero() {
			g.FirstTimestamp = timestamppb.New(now.Add(time.Duration(index) * time.Millisecond))
		}
		newEventOrder = append(newEventOrder, g)
	}
	sort.Sort(newEventOrder)

	// get before last events
	lastEvent := getLastEvent(ordered)
	for _, ev := range newEventOrder {
		// if lastEvent was empty, mean ordered was empty
		// lastEvent assignment this events
		if lastEvent == nil {
			ordered = append(ordered, ev)
			lastEvent = ev
			continue
		}

		// get last events key and this events key
		lastEventKey := makeEventGroupKey(lastEvent)
		thsEventKey := makeEventGroupKey(ev)

		// same like
		if strings.EqualFold(lastEventKey, thsEventKey) {
			// compare time and exchange
			if !ev.FirstTimestamp.AsTime().IsZero() && ev.FirstTimestamp.AsTime().Before(lastEvent.FirstTimestamp.AsTime()) {
				lastEvent.FirstTimestamp = ev.FirstTimestamp
			}
			if ev.LastTimestamp.AsTime().After(lastEvent.LastTimestamp.AsTime()) {
				lastEvent.LastTimestamp = ev.LastTimestamp
			}
			// count++
			// lastEvent not change
			lastEvent.Count++
			continue
		} else {
			// not like,
			// append this events to orders
			// lastEvent assignment this events
			ordered = append(ordered, ev)
			lastEvent = ev
		}
	}
	return ordered
}

func (o orderedEvents) Len() int { return len(o) }
func (o orderedEvents) Less(i, j int) bool {
	return o[i].LastTimestamp.AsTime().Before(o[j].LastTimestamp.AsTime())
}
func (o orderedEvents) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

func makeEventGroupKey(se *basepb.PipelineEvent) string {
	return fmt.Sprintf("%s:%s:%s", se.Source.Component, se.Reason, se.Message)
}

func makeEventReportMeta(events []*basepb.PipelineEvent) apistructs.PipelineReportMeta {
	return apistructs.PipelineReportMeta{apistructs.PipelineReportEventMetaKey: events}
}
