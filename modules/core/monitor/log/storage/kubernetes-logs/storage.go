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

package kuberneteslogs

import (
	"context"
	"errors"
	"io"
	"math"
	"os"
	"regexp"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/modules/core/monitor/log/storage"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
)

type queryFunc func(it *logsIterator, opts *v1.PodLogOptions) (io.ReadCloser, error)

type cStorage struct {
	log          logs.Logger
	getQueryFunc func(clusterName string) (func(it *logsIterator, opts *v1.PodLogOptions) (io.ReadCloser, error), error)
	pods         PodInfoQueryer
	bufferLines  int64
	timeSpan     int64
}

var _ storage.Storage = (*cStorage)(nil)

func (s *cStorage) NewWriter(ctx context.Context) (storekit.BatchWriter, error) {
	return nil, storekit.ErrOpNotSupported
}

func newPodLogOptions(container_name string, startTime int64) *v1.PodLogOptions {
	var sinceTime *metav1.Time
	if startTime > 0 {
		sinceTime = &metav1.Time{Time: time.Unix(startTime/int64(time.Second), startTime%int64(time.Second))}
	}
	return &v1.PodLogOptions{
		Container:                    container_name,
		Follow:                       false,
		Previous:                     false,
		SinceSeconds:                 nil,
		SinceTime:                    sinceTime,
		Timestamps:                   true,
		TailLines:                    nil,
		LimitBytes:                   nil,
		InsecureSkipTLSVerifyBackend: true,
	}
}

var podInfoKeys = []string{"cluster_name", "pod_namespace", "pod_name", "container_name"}

const (
	defaultBufferLines = 1024
	defaultTimeSpan    = 3 * time.Minute
)

func (s *cStorage) Iterator(ctx context.Context, sel *storage.Selector) (storekit.Iterator, error) {
	var err error
	var id string
	matcher := func(data *pb.LogItem, it *logsIterator) bool { return true }
	for _, filter := range sel.Filters {
		switch filter.Key {
		case "id":
			if filter.Op != storage.EQ {
				s.log.Debugf("id  only support EQ filter, ingore kubernetes logs query")
				return storekit.EmptyIterator{}, nil
			}
			if len(id) <= 0 {
				id, _ = filter.Value.(string)
			}
		case "source":
			if filter.Op != storage.EQ {
				s.log.Debugf("source only support EQ filter, ingore kubernetes logs query")
				return storekit.EmptyIterator{}, nil
			}
			source, _ := filter.Value.(string)
			if len(source) > 0 && source != "container" {
				return storekit.EmptyIterator{}, nil
			}
		case "stream":
			// ingore
		case "content":
			switch filter.Op {
			case storage.REGEXP:
				exp, _ := filter.Value.(string)
				if len(exp) > 0 {
					regex, err := regexp.Compile(exp)
					if err != nil {
						s.log.Debugf("invalid regexp %q, ingore kubernetes logs query", exp)
						return storekit.EmptyIterator{}, nil
					}
					matcher = func(data *pb.LogItem, it *logsIterator) bool {
						return regex.MatchString(data.Content)
					}
				}
			case storage.EQ:
				val, _ := filter.Value.(string)
				if len(val) > 0 {
					matcher = func(data *pb.LogItem, it *logsIterator) bool {
						return data.Content == val
					}
				}
			}
		}
	}
	if len(id) <= 0 {
		s.log.Debugf("not found id, ingore kubernetes logs query")
		return storekit.EmptyIterator{}, nil
	}

	tags := make(map[string]string)
	for _, key := range podInfoKeys {
		value, _ := sel.Options[key].(string)
		if len(value) > 0 {
			tags[key] = value
		}
	}
	for _, key := range podInfoKeys {
		if len(tags[key]) <= 0 {
			info, err := s.pods.GetPodInfo(id, sel)
			if err != nil {
				s.log.Debugf("failed to query pod info for container(%q): %s, ingore kubernetes logs query", id, err)
				return storekit.EmptyIterator{}, nil
			}
			for k, v := range info {
				if len(tags[k]) <= 0 {
					tags[k] = v
				}
			}
			break
		}
	}
	for _, key := range podInfoKeys {
		if len(tags[key]) <= 0 {
			s.log.Debugf("not found %q for container(%q), ingore kubernetes logs query", key, id)
			return storekit.EmptyIterator{}, nil
		}
	}
	clusterName := tags["cluster_name"]

	queryFunc, err := s.getQueryFunc(clusterName)
	if err != nil {
		s.log.Debugf("failed to GetClient(%q): %s, ingore kubernetes logs query", clusterName, err)
		return storekit.EmptyIterator{}, nil
	}

	bufferLines := s.bufferLines
	if bufferLines <= 0 {
		bufferLines = defaultBufferLines
	}
	timeSpan := s.timeSpan
	if timeSpan <= 0 {
		bufferLines = int64(defaultTimeSpan)
	}
	return &logsIterator{
		ctx:           ctx,
		sel:           sel,
		id:            id,
		podNamespace:  tags["pod_namespace"],
		podName:       tags["pod_name"],
		containerName: tags["container_name"],
		matcher:       matcher,
		queryFunc:     queryFunc,
		pageSize:      bufferLines,
		timeSpan:      int64(3 * time.Minute),
	}, nil
}

type iteratorDir int8

const (
	iteratorInitial = iota
	iteratorForward
	iteratorBackward
)

type logsIterator struct {
	ctx           context.Context
	sel           *storage.Selector
	id            string
	podNamespace  string
	podName       string
	containerName string
	matcher       func(data *pb.LogItem, it *logsIterator) bool

	queryFunc queryFunc
	pageSize  int64
	timeSpan  int64
	dir       iteratorDir

	lastStartTime int64
	lastEndTime   int64
	lastStartLine string
	lastEndLine   string
	buffer        []*pb.LogItem
	offset        int
	value         *pb.LogItem
	err           error
	closed        bool
}

func (it *logsIterator) First() bool {
	if it.checkClosed() && it.Error() != nil {
		return false
	}
	it.dir = iteratorForward
	it.fetch(it.sel.Start, it.pageSize, false)
	if it.offset < len(it.buffer) {
		it.value = it.buffer[it.offset]
		it.offset++
		return true
	}
	return false
}

func (it *logsIterator) Last() bool {
	if it.checkClosed() && it.Error() != nil {
		return false
	}
	it.dir = iteratorBackward
	startTime := it.sel.End - it.timeSpan
	if startTime < it.sel.Start {
		startTime = it.sel.Start
	}
	it.fetch(startTime, -1, true)
	if it.offset >= 0 && it.offset < len(it.buffer) {
		it.value = it.buffer[it.offset]
		it.offset--
		return true
	}
	return false
}

func (it *logsIterator) Next() bool {
	if it.checkClosed() {
		return false
	}
	if it.dir == iteratorBackward {
		it.err = storekit.ErrOpNotSupported
		return false
	}
	it.dir = iteratorForward
	if it.offset >= 0 && it.offset < len(it.buffer) {
		it.value = it.buffer[it.offset]
		it.offset++
		return true
	}
	if it.err != nil {
		return false
	}
	if it.lastEndTime != 0 {
		if it.lastEndTime <= it.sel.End {
			it.fetch(it.lastEndTime, it.pageSize, false)
		}
	} else {
		it.fetch(it.sel.Start, it.pageSize, false)
	}
	if it.offset >= 0 && it.offset < len(it.buffer) {
		it.value = it.buffer[it.offset]
		it.offset++
		return true
	}
	return false
}

func (it *logsIterator) Prev() bool {
	if it.checkClosed() {
		return false
	}
	if it.dir == iteratorForward {
		it.err = storekit.ErrOpNotSupported
		return false
	}
	it.dir = iteratorBackward
	if it.offset >= 0 && it.offset < len(it.buffer) {
		it.value = it.buffer[it.offset]
		it.offset--
		return true
	}
	if it.err != nil {
		return false
	}
	if it.lastStartTime != 0 {
		if it.lastStartTime > it.sel.Start {
			startTime := it.lastStartTime - it.timeSpan
			if startTime < it.sel.Start {
				startTime = it.sel.Start
			}
			it.fetch(startTime, -1, true)
		}
	} else {
		startTime := it.sel.End - it.timeSpan
		if startTime < it.sel.Start {
			startTime = it.sel.Start
		}
		it.fetch(startTime, -1, true)
	}
	if it.offset >= 0 && it.offset < len(it.buffer) {
		it.value = it.buffer[it.offset]
		it.offset--
		return true
	}
	return false
}

func (it *logsIterator) Value() interface{} { return it.value }
func (it *logsIterator) Error() error {
	if it.err == io.EOF {
		return nil
	}
	return it.err
}

func (it *logsIterator) Close() error {
	it.closed = true
	return nil
}

func (it *logsIterator) checkClosed() bool {
	if it.closed {
		if it.err == nil {
			it.err = storekit.ErrIteratorClosed
		}
		return true
	}
	select {
	case <-it.ctx.Done():
		if it.err == nil {
			it.err = storekit.ErrIteratorClosed
		}
		return true
	default:
	}
	return false
}

var stdout2 = os.Stdout

var errLimited = errors.New("limited")

func (it *logsIterator) fetch(startTime, limit int64, backward bool) {
	it.buffer = nil
	endTime := it.sel.End
	for it.err == nil && len(it.buffer) <= 0 && startTime >= it.sel.Start {
		func() error {
			opts := newPodLogOptions(it.containerName, startTime)
			reader, err := it.queryFunc(it, opts)
			if err != nil {
				it.err = err
				return it.err
			}
			defer reader.Close()
			var (
				minTime, maxTime int64 = math.MaxInt64, 0
				first            bool  = true
				firstLine        string
			)
			err = parseLines(reader, func(line []byte) error {
				text := string(line)
				if first {
					first = false
					firstLine = text
				}
				data, content := parseLine(text, it)
				if data.Timestamp > 0 {
					if data.Timestamp < minTime {
						minTime = data.Timestamp
					}
					if data.Timestamp > maxTime {
						maxTime = data.Timestamp
					}
					if backward {
						if data.Timestamp < it.sel.Start {
							return io.EOF
						}
						if endTime <= data.Timestamp && it.lastStartLine != text {
							if startTime <= it.sel.Start {
								return io.EOF
							}
							return errLimited
						}
					} else {
						if data.Timestamp < it.sel.Start {
							// continue
							return nil
						}
						if it.sel.End <= data.Timestamp {
							return io.EOF
						}
					}
					if int(limit) > 0 && len(it.buffer) >= int(limit) &&
						data.Timestamp != it.buffer[len(it.buffer)-1].Timestamp {
						return errLimited
					}
				}
				parseContent(content, data)
				if it.lastEndLine == text {
					// remove duplicate
					return nil
				}
				it.lastEndLine = text

				if !it.matcher(data, it) {
					return nil
				}
				it.buffer = append(it.buffer, data)
				return nil
			})
			it.lastStartLine = firstLine
			it.lastStartTime = minTime
			if startTime < it.lastStartTime {
				it.lastStartTime = startTime
			}
			if maxTime != 0 && (it.lastEndTime == 0 || !backward) {
				it.lastEndTime = maxTime
			}
			if backward {
				startTime = it.lastStartTime - it.timeSpan
				if startTime < it.sel.Start {
					startTime = it.sel.Start
				}
				if maxTime != 0 {
					endTime = maxTime
				}
			}
			if err != nil && err != errLimited {
				it.err = err
				return it.err
			}
			return nil
		}()
	}
	if backward {
		it.offset = len(it.buffer) - 1
	} else {
		it.offset = 0
	}
}
