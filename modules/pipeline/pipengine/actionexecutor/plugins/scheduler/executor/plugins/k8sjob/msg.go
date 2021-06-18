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

package k8sjob

import (
	"context"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Message Event message
type Message struct {
	Timestamp int64
	Reason    string
	Message   string
	Comment   string
}

// MessageList 事件消息列表
type MessageList []Message

func (em MessageList) Len() int           { return len(em) }
func (em MessageList) Swap(i, j int)      { em[i], em[j] = em[j], em[i] }
func (em MessageList) Less(i, j int) bool { return em[i].Timestamp < em[j].Timestamp }

func (k *K8sJob) getLastMsg(ctx context.Context, namespace, name string) (lastMsg string, err error) {
	var ems MessageList

	eventList, err := k.client.ClientSet.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}

	for i := range eventList.Items {
		e := &eventList.Items[i]
		if e.InvolvedObject.Kind != "Pod" || !strings.HasPrefix(e.InvolvedObject.Name, name) {
			continue
		}

		// One-stage analysis, the reasons are intuitively visible
		if cmt, ok := interestedEventCommentFirstMap[e.Reason]; ok {
			ems = append(ems, Message{
				Timestamp: e.LastTimestamp.Unix(),
				Reason:    e.Reason,
				Message:   e.Message,
				Comment:   cmt,
			})
			continue
		}

		// Two-stage analysis requires event.message analysis
		cmt, err := secondAnalyzePodEventComment(e.Reason, e.Message)
		if err == nil {
			ems = append(ems, Message{
				Timestamp: e.LastTimestamp.Unix(),
				Reason:    e.Reason,
				Message:   e.Message,
				Comment:   cmt,
			})
		}
	}

	sort.Sort(ems)
	if len(ems) > 0 {
		lastMsg = ems[len(ems)-1].Comment
	}
	return lastMsg, nil
}
