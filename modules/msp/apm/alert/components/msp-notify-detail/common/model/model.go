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

package model

type EventStatusInfo struct {
	Type  string              `json:"type,omitempty"`
	Data  EventStatusInfoData `json:"data"`
	Props Props               `json:"props"`
}

type EventStatusInfoData struct {
	Data Data `json:"data"`
}

type Data struct {
	NotificationMethod string       `json:"notificationMethod"`
	SendStatus         []SendStatus `json:"sendStatus"`
}

type SendStatus struct {
	Label                        string `json:""label`
	Color                        string `json:"color"`
	SendTime                     string `json:"sendTime"`
	AssociatedNotificationGroups string `json:"associatedNotificationGroups"`
	AssociatedAlarmRule          string `json:"associatedAlarmRule"`
}

type Props struct {
}
