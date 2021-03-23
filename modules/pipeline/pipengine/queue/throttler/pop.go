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
