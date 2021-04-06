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

package events

type Hook interface {
	HandleWebhook
	HandleWebSocket
	HandleDingDing
	HandleHTTP
}

type HandleWebhook interface{ HandleWebhook() error }
type HandleWebSocket interface{ HandleWebSocket() error }
type HandleDingDing interface{ HandleDingDing() error }
type HandleHTTP interface{ HandleHTTP() error }

type HookType string

const (
	HookTypeWebHook   HookType = "WEBHOOK"
	HookTypeWebSocket HookType = "WEBSOCKET"
	HookTypeDINGDING  HookType = "DINGDING"
	HookTypeHTTP      HookType = "HTTP"
)
