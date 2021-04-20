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

package throttler

import (
	"time"
)

type AddKeyToQueueRequest struct {
	QueueName    string
	QueueWindow  *int64
	Priority     int64
	CreationTime time.Time
}

type PopDetail struct {
	QueueName string
	CanPop    bool
	Reason    string
}

func newPopDetail(qName string, canPop bool, reason string) PopDetail {
	return PopDetail{QueueName: qName, CanPop: canPop, Reason: reason}
}
