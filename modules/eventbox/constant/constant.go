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

package constant

import (
	"path/filepath"
)

var (
	EventboxDir = "/eventbox"
	MessageDir  = filepath.Join(EventboxDir, "messages")

	// schemon input
	// ScheMonServiceLockKey  = filepath.Join(EventboxDir, "schemonservicelock")
	// ScheMonServiceWatchDir = filepath.Join("/dice/service")
	// ScheMonServiceQueryURL = "/v1/notify/runtime/%s/%s" // /v1/notify/runtime/<namespace>/<name>
	ScheMonJobLockKey  = filepath.Join(EventboxDir, "schemonjoblock")
	ScheMonJobWatchDir = filepath.Join("/dice/job/")
	ScheMonJobQueryURL = "/v1/notify/job/%s/%s" // /v1/notify/job/<namespace>/<name>

	// register
	RegisterDir      = filepath.Join(EventboxDir, "register")
	RegisterLabelKey = "/REGISTERED_LABEL"

	// webhook
	WebhookLabelKey = "/WEBHOOK"
	WebhookDir      = filepath.Join(EventboxDir, "webhook")
)
