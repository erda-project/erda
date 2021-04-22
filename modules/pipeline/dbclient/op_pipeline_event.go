// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package dbclient

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// GetPipelineEvents get pipeline events from reports.
// return: report, events, error
func (client *Client) GetPipelineEvents(pipelineID uint64, ops ...SessionOption) (*spec.PipelineReport, []*apistructs.PipelineEvent, error) {
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
	var events []*apistructs.PipelineEvent
	if err := json.Unmarshal(b, &events); err != nil {
		return &report, nil, nil
	}
	return &report, events, nil
}

func (client *Client) AppendPipelineEvent(pipelineID uint64, newEvents []*apistructs.PipelineEvent, ops ...SessionOption) error {
	if len(newEvents) == 0 {
		return nil
	}

	session := client.NewSession(ops...)
	defer session.Close()

	// get latest events
	report, events, err := client.GetPipelineEvents(pipelineID, ops...)
	if err != nil {
		return err
	}
	// merge all events
	events = append(events, newEvents...)
	// group events by event detail
	group := make(map[string][]*apistructs.PipelineEvent)
	for _, event := range events {
		key := makeEventGroupKey(event)
		group[key] = append(group[key], event)
	}
	mergedGroup := make(map[string]*apistructs.PipelineEvent)
	for key, ses := range group {
		newSe := *ses[0]
		for i, se := range ses {
			if i == 0 {
				continue
			}
			if !se.FirstTimestamp.IsZero() && se.FirstTimestamp.Before(newSe.FirstTimestamp) {
				newSe.FirstTimestamp = se.FirstTimestamp
			}
			if se.LastTimestamp.After(newSe.LastTimestamp) {
				newSe.LastTimestamp = se.LastTimestamp
			}
			newSe.Count++
		}
		// set default
		now := time.Now()
		if newSe.FirstTimestamp.IsZero() {
			newSe.FirstTimestamp = now
		}
		if newSe.LastTimestamp.IsZero() {
			newSe.FirstTimestamp = now
		}
		// add to merged group
		mergedGroup[key] = &newSe
	}
	// order message by firstTimestamp
	var ordered orderedEvents
	for _, g := range mergedGroup {
		ordered = append(ordered, g)
	}
	sort.Sort(ordered)
	if len(ordered) == 0 {
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

type orderedEvents []*apistructs.PipelineEvent

func (o orderedEvents) Len() int           { return len(o) }
func (o orderedEvents) Less(i, j int) bool { return o[i].FirstTimestamp.Before(o[j].FirstTimestamp) }
func (o orderedEvents) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }

func makeEventGroupKey(se *apistructs.PipelineEvent) string {
	return fmt.Sprintf("%s:%s:%s", se.Source.Component, se.Reason, se.Message)
}

func makeEventReportMeta(events []*apistructs.PipelineEvent) apistructs.PipelineReportMeta {
	return apistructs.PipelineReportMeta{apistructs.PipelineReportEventMetaKey: events}
}
