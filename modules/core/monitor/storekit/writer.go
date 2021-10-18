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

package storekit

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// WrapBatchWriter .
func WrapBatchWriter(
	bw BatchWriter,
	bufferSize uint64, // buffer size
	timeout time.Duration, // timeout for buffer flush
	errorh func(error) error, // error handler
) Writer {
	w := &channelWriter{
		dataCh:  make(chan Data, bufferSize),
		closeCh: make(chan error, 1),
	}
	go w.run(bw, int(bufferSize), timeout, errorh)
	return w
}

type channelWriter struct {
	dataCh  chan Data
	closeCh chan error
}

func (w *channelWriter) Write(data Data) error {
	w.dataCh <- data
	return nil
}

func (w *channelWriter) WriteN(data ...Data) (int, error) {
	for _, item := range data {
		w.dataCh <- item
	}
	return len(data), nil
}

func (w *channelWriter) Close() error {
	close(w.dataCh)
	return <-w.closeCh
}

func (w *channelWriter) run(bw BatchWriter, capacity int, timeout time.Duration, errorh func(error) error) {
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

// Stdout .
type Stdout struct {
	Out io.Writer
	Enc *json.Encoder
}

// DefaultStdout .
var DefaultStdout = Stdout{
	Out: os.Stdout,
	Enc: json.NewEncoder(os.Stdout),
}

func init() {
	DefaultStdout.Enc.SetIndent("", "\t")
}

func (w Stdout) Write(val Data) error {
	w.Enc.Encode(val)
	return nil
}

func (w Stdout) WriteN(vals ...Data) (int, error) {
	for _, val := range vals {
		w.Enc.Encode(val)
	}
	return len(vals), nil
}

func (w Stdout) Close() error { return nil }
