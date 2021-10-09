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

import "time"

// ErrorHandler .
type WriteErrorHandler func(error) error

// WrapBatchWriter .
func WrapBatchWriter(
	bw BatchWriter,
	bufferSize uint64, // buffer size
	timeout time.Duration, // timeout for buffer flush
	errorh WriteErrorHandler, // error handler
) Writer {
	w := &channelWriter{
		dataCh:  make(chan *Data, bufferSize),
		closeCh: make(chan error, 1),
	}
	go w.run(bw, int(bufferSize), timeout, errorh)
	return w
}

type channelWriter struct {
	dataCh  chan *Data
	closeCh chan error
}

func (w *channelWriter) Write(data *Data) error {
	w.dataCh <- data
	return nil
}

func (w *channelWriter) WriteN(data ...*Data) (int, error) {
	for _, item := range data {
		w.dataCh <- item
	}
	return len(data), nil
}

func (w *channelWriter) Close() error {
	close(w.dataCh)
	return <-w.closeCh
}

func (w *channelWriter) run(bw BatchWriter, capacity int, timeout time.Duration, errorh WriteErrorHandler) {
	buf := NewBufferedWriter(bw, capacity)
	tick := time.NewTicker(timeout)
	var err error
	defer func() {
		tick.Stop()
		cerr := buf.Close()
		if cerr != nil {
			cerr = errorh(cerr)
			if err == nil {
				err = cerr
			}
		}
		w.closeCh <- err
	}()
	for {
		select {
		case data, ok := <-w.dataCh:
			if !ok {
				return
			}
			err = buf.Write(data)
			if err != nil {
				if errorh != nil {
					err = errorh(err)
					if err != nil {
						return
					}
				} else {
					return
				}
			}
		case <-tick.C:
			err = buf.Flush()
			if err != nil {
				if errorh != nil {
					err = errorh(err)
					if err != nil {
						return
					}
				} else {
					return
				}
			}
		}
	}
}
