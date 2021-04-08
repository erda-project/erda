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

// Package event manipulates the k8s api of event object
package event

import (
	"context"
	"fmt"
	"sort"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/kubelet/events"
)

// Message 事件消息
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

// AnalyzePodEvents 分析 pod events, 转成可解读的内容
func (e *Event) AnalyzePodEvents(namespace, name string) (MessageList, error) {
	var ems MessageList

	var (
		eventList *apiv1.EventList
		err       error
	)
	if e.k8sClient != nil {
		eventList, err = e.k8sClient.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{})
	} else {
		eventList, err = e.ListByNamespace(namespace)
	}
	if err != nil {
		return nil, errors.Errorf("failed to list pod events, namespace: %s, podName: %s, (%v)",
			namespace, name, err)
	}

	for i := range eventList.Items {
		e := &eventList.Items[i]
		if e.InvolvedObject.Kind != "Pod" || !strings.HasPrefix(e.InvolvedObject.Name, name) {
			continue
		}

		// 一阶段分析，原因直观可见
		if cmt, ok := interestedEventCommentFirstMap[e.Reason]; ok {
			ems = append(ems, Message{
				Timestamp: e.LastTimestamp.Unix(),
				Reason:    e.Reason,
				Message:   e.Message,
				Comment:   cmt,
			})
			continue
		}

		// 二阶段分析，需要进行 event.message 解析
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
	return ems, nil
}

// 一阶段信息获取，可直观分析出原因
var interestedEventCommentFirstMap = map[string]string{
	// events.FailedToInspectImage:     errInspectImage,
	events.ErrImageNeverPullPolicy:  errImageNeverPull,
	events.NetworkNotReady:          errNetworkNotReady,
	events.FailedAttachVolume:       errMountVolume,
	events.FailedMountVolume:        errMountVolume,
	events.VolumeResizeFailed:       errMountVolume,
	events.FileSystemResizeFailed:   errMountVolume,
	events.FailedMapVolume:          errMountVolume,
	events.WarnAlreadyMountedVolume: errAlreadyMountedVolume,
	events.NodeRebooted:             errNodeRebooted,
}

// 二阶段分析，需要解析 event.message
func secondAnalyzePodEventComment(reason, message string) (string, error) {
	switch reason {
	case "FailedScheduling":
		return parseFailedScheduling(message)
	case events.FailedToPullImage:
		return parseFailedReason(message)
	default:
		// TODO: 补充更多的 reason
		return "", errors.Errorf("invalid event reason: %s", reason)
	}
}

// 解析调度失败的原因
func parseFailedScheduling(message string) (string, error) {
	var (
		totalNodes    int
		notMatchNodes int
		tmpStr        string
	)

	splitFunc := func(r rune) bool {
		return r == ':' || r == ','
	}
	msgSlice := strings.FieldsFunc(message, splitFunc)

	// 节点资源不可用
	// 可以分为：
	// 1. 标签不匹配。 示例："0/8 nodes are available: 8 node(s) didn't match node selector."
	// 2. CPU 不足。示例："0/8 nodes are available: 5 Insufficient cpu, 5 node(s) didn't match node selector."
	// 3. 内存不足。示例："0/8 nodes are available: 5 node(s) didn't match node selector, 5 Insufficient memory."
	if strings.Contains(message, "nodes are available") {
		_, err := fmt.Sscanf(message, "0/%d nodes are available: %s", &totalNodes, &tmpStr)
		if err != nil {
			return "", errors.Errorf("failed to parse totalNodes num, body: %s, (%v)", message, err)
		}

		for _, msg := range msgSlice {
			if strings.Contains(msg, "node(s) didn't match node selector") {
				_, err := fmt.Sscanf(msg, "%d node(s) didn't match node selector", &notMatchNodes)
				if err != nil {
					return "", errors.Errorf("failed to parse notMatchNodes num, body: %s, (%v)", msg, err)
				}
			}
		}

		if totalNodes > 0 && (totalNodes == notMatchNodes) {
			return errNodeSelectorMismatching, nil
		}

		if strings.Contains(message, "Insufficient cpu") {
			return errInsufficientFreeCPU, nil
		}

		if strings.Contains(message, "Insufficient memory") {
			return errInsufficientFreeMemory, nil
		}
	}

	// TODO: 补充更多的信息
	return "", errors.New("unexpected")
}

// 解析运行失败的原因
func parseFailedReason(message string) (string, error) {
	switch {
	// 无效的镜像名
	case strings.Contains(message, "InvalidImageName"):
		return errInvalidImageName, nil
	// 无效的镜像
	case strings.Contains(message, "ImagePullBackOff"):
		return errPullImage, nil
	default:
		// TODO: 分析更多的原因
		return "", errors.New("unexpected")
	}
}
