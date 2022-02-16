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

package common

import (
	"time"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
)

func SetAlertEventToGlobalState(gs cptype.GlobalStateData, alertEvent *pb.AlertEventItem) {
	gs[StateKeyAlertEvent] = alertEvent
	gs[StateKeyPageTitle] = alertEvent.Name
}

func GetAlertEventFromGlobalState(gs cptype.GlobalStateData) *pb.AlertEventItem {
	item, ok := gs[StateKeyAlertEvent]
	if !ok {
		return nil
	}

	typedItem, ok := item.(*pb.AlertEventItem)
	if !ok {
		return nil
	}

	return typedItem
}

func FormatTimeMs(timestamp int64) string {
	if timestamp == 0 {
		return "-"
	}

	return time.Unix(timestamp/1e3, 0).Format(TimeFormatLayout)
}
