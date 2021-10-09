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

package storage

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type (
	// Data .
	Data struct {
		Timestamp int64                  `json:"timestamp"`
		Labels    map[string]string      `json:"labels"`
		Fields    map[string]interface{} `json:"fields"`
	}

	// StorageReader .
	StorageReader interface {
		Query(ctx context.Context, sel *Selector) (*QueryResult, error)
		Iterator(ctx context.Context, sel *Selector) Iterator
	}

	// StorageWriter .
	StorageWriter interface {
		NewWriter(opts *WriteOptions) (Writer, error)
	}

	// StorageBatchWriter .
	StorageBatchWriter interface {
		NewBatchWriter(opts *WriteOptions) (BatchWriter, error)
	}

	// Storage .
	Storage interface {
		StorageReader
		StorageWriter
		Mode() Mode
	}

	// Writer .
	Writer interface {
		Write(val *Data) error
		Close() error
	}

	// BatchWriter .
	BatchWriter interface {
		WriteN(vals ...*Data) (int, error)
		Close() error
	}

	// Syncer
	Syncer interface {
		Sync() error
	}

	// Selector .
	Selector struct {
		StartTime     int64
		EndTime       int64
		Limit         int64
		PartitionKeys []string
		Labels        map[string]string
		Matcher       Matcher
		// Comparer      Comparer
	}

	// Matcher .
	Matcher interface {
		Match(val *Data) bool
	}

	// Comparer
	// Comparer interface {
	// 	Compare(a, b Data) int
	// }

	// QueryResult .
	QueryResult struct {
		OriginalTotal int64
		Values        []*Data
	}

	// Iterator .
	Iterator interface {
		First() bool
		Last() bool
		Next() bool
		Prev() bool
		Value() *Data
		Error() error
		Close() error
	}
)

type (
	// WriteOption .
	WriteOption func(opts *WriteOptions)
	// WriteOptions
	WriteOptions struct {
		TTL              time.Duration
		Timeout          time.Duration
		PartitionKeyFunc func(val *Data) string
		KeyFunc          func(val *Data) string
	}
)

// Mode .
type Mode string

const (
	// ModeReadonly .
	ModeReadonly Mode = "ro"
	// ModeReadWrite .
	ModeReadWrite Mode = "rw"
)

// WithTTL .
func WithTTL(ttl time.Duration) WriteOption {
	return func(opts *WriteOptions) {
		opts.TTL = ttl
	}
}

// WithWriteTimeout .
func WithWriteTimeout(timeout time.Duration) WriteOption {
	return func(opts *WriteOptions) {
		opts.Timeout = timeout
	}
}

// WithPartitionKeyFunc .
func WithPartitionKeyFunc(pk func(val *Data) string) WriteOption {
	return func(opts *WriteOptions) {
		opts.PartitionKeyFunc = pk
	}
}

// WithKeyFunc .
func WithKeyFunc(pk func(val *Data) string) WriteOption {
	return func(opts *WriteOptions) {
		opts.KeyFunc = pk
	}
}

// NewWriteOptions .
func NewWriteOptions(opts ...WriteOption) *WriteOptions {
	o := &WriteOptions{}
	for _, opt := range opts {
		opt(o)
	}
	if o.KeyFunc == nil {
		o.KeyFunc = func(val *Data) string { return "" }
	}
	if o.PartitionKeyFunc == nil {
		o.PartitionKeyFunc = func(val *Data) string { return "" }
	}
	return o
}

// WriteError .
type WriteError struct {
	Data *Data
	Err  error
}

func (e *WriteError) Error() string {
	return e.Err.Error()
}

// BatchWriteError .
type BatchWriteError struct {
	Errors []*WriteError
}

func (e *BatchWriteError) Error() string {
	return fmt.Sprintf("bulk writes occur errors(%d)", len(e.Errors))
}

var (
	// ErrInvalidData .
	ErrInvalidData = errors.New("invalid data")
	// ErrEmptyData .
	ErrEmptyData = errors.New("empty data")
	// ErrIteratorClosed .
	ErrIteratorClosed = errors.New("iterator closed")
	// ErrOpNotSupported .
	ErrOpNotSupported = errors.New("operation not supported")
)

const (
	// DefaultLimit .
	DefaultLimit = 1000
)
