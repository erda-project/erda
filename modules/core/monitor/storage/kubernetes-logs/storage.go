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
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/erda-project/erda/modules/core/monitor/storage"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type queryFunc func(ctx context.Context, namespace, pod string, opts *v1.PodLogOptions) (io.ReadCloser, error)

type cStorage struct {
	queryFunc   queryFunc
	bufferLines int64
}

var _ storage.Storage = (*cStorage)(nil)

const defaultBufferLines = 2 * 1024

func (s *cStorage) Mode() storage.Mode { return storage.ModeReadonly }

func (s *cStorage) Write(val *storage.Data, opts *storage.WriteOptions) error {
	return storage.ErrOpNotSupported
}

func (s *cStorage) NewWriter(opts *storage.WriteOptions) (storage.Writer, error) {
	return nil, storage.ErrOpNotSupported
}

func (s *cStorage) Query(ctx context.Context, sel *storage.Selector) (*storage.QueryResult, error) {
	if len(sel.PartitionKeys) != 3 {
		return nil, fmt.Errorf("invalid partition keys")
	}
	opts := newPodLogOptions(sel.PartitionKeys[2], sel.StartTime)
	reader, err := s.queryFunc(ctx, sel.PartitionKeys[0], sel.PartitionKeys[1], opts)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	result := &storage.QueryResult{}
	limit := int(sel.Limit)
	if err = parseLines(reader, func(line []byte) error {
		if limit > 0 && len(result.Values) >= limit {
			return io.EOF
		}
		result.OriginalTotal++
		data, content := parseLine(string(line), sel)
		if data.Timestamp > 0 {
			if data.Timestamp < sel.StartTime {
				return nil
			} else if sel.EndTime <= data.Timestamp {
				return io.EOF
			}
		}
		parseContent(content, data)
		if !matchData(data, sel) {
			return nil
		}
		result.Values = append(result.Values, data)
		return nil
	}); err != nil && err != io.EOF {
		return nil, err
	}
	return result, nil
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

func matchLabels(labels, match map[string]string) bool {
	for k, v := range match {
		if val, ok := labels[k]; !ok || v != val {
			return false
		}
	}
	return true
}

func matchData(data *storage.Data, sel *storage.Selector) bool {
	if matchLabels(data.Labels, sel.Labels) {
		if sel.Matcher != nil {
			return sel.Matcher.Match(data)
		}
		return true
	}
	return false
}

func (s *cStorage) Iterator(ctx context.Context, sel *storage.Selector) storage.Iterator {
	var err error
	if len(sel.PartitionKeys) != 3 {
		err = fmt.Errorf("invalid partition keys")
	}
	return &logsIterator{
		ctx:       ctx,
		sel:       sel,
		queryFunc: s.queryFunc,
		pageSize:  s.bufferLines,
		timeSpan:  int64(3 * time.Minute),
		err:       err,
	}
}

type iteratorDir int8

const (
	iteratorInitial = iota
	iteratorForward
	iteratorBackward
)

type logsIterator struct {
	ctx       context.Context
	sel       *storage.Selector
	queryFunc queryFunc
	pageSize  int64
	timeSpan  int64
	dir       iteratorDir

	lastStartTime int64
	lastEndTime   int64
	lastStartLine string
	lastEndLine   string
	buffer        []*storage.Data
	offset        int
	value         *storage.Data
	err           error
	closed        bool
}

func (it *logsIterator) First() bool {
	if it.checkClosed() && it.Error() != nil {
		return false
	}
	it.dir = iteratorForward
	it.fetch(it.sel.StartTime, it.pageSize, false)
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
	startTime := it.sel.EndTime - it.timeSpan
	if startTime < it.sel.StartTime {
		startTime = it.sel.StartTime
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
		it.err = storage.ErrOpNotSupported
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
		if it.lastEndTime <= it.sel.EndTime {
			it.fetch(it.lastEndTime, it.pageSize, false)
		}
	} else {
		it.fetch(it.sel.StartTime, it.pageSize, false)
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
		it.err = storage.ErrOpNotSupported
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
		if it.lastStartTime > it.sel.StartTime {
			startTime := it.lastStartTime - it.timeSpan
			if startTime < it.sel.StartTime {
				startTime = it.sel.StartTime
			}
			it.fetch(startTime, -1, true)
		}
	} else {
		startTime := it.sel.EndTime - it.timeSpan
		if startTime < it.sel.StartTime {
			startTime = it.sel.StartTime
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

func (it *logsIterator) Value() *storage.Data { return it.value }
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
			it.err = storage.ErrIteratorClosed
		}
		return true
	}
	return false
}

var stdout2 = os.Stdout

var errLimited = errors.New("limited")

func (it *logsIterator) fetch(startTime, limit int64, backward bool) {
	it.buffer = nil
	endTime := it.sel.EndTime
	for it.err == nil && len(it.buffer) <= 0 && startTime >= it.sel.StartTime {
		func() error {
			opts := newPodLogOptions(it.sel.PartitionKeys[2], startTime)
			reader, err := it.queryFunc(it.ctx, it.sel.PartitionKeys[0], it.sel.PartitionKeys[1], opts)
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
				data, content := parseLine(text, it.sel)
				if data.Timestamp > 0 {
					if data.Timestamp < minTime {
						minTime = data.Timestamp
					}
					if data.Timestamp > maxTime {
						maxTime = data.Timestamp
					}
					if backward {
						if data.Timestamp < it.sel.StartTime {
							return io.EOF
						}
						if endTime <= data.Timestamp && it.lastStartLine != text {
							if startTime <= it.sel.StartTime {
								return io.EOF
							}
							return errLimited
						}
					} else {
						if data.Timestamp < it.sel.StartTime {
							// continue
							return nil
						}
						if it.sel.EndTime <= data.Timestamp {
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

				if !matchData(data, it.sel) {
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
				if startTime < it.sel.StartTime {
					startTime = it.sel.StartTime
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
