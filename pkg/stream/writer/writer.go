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

package writer

import (
	"time"

	"github.com/recallsong/go-utils/errorx"
)

// Writer .
type Writer interface {
	Write(data interface{}) error
	WriteN(data ...interface{}) (int, error)
	Close() error
}

// ErrorHandler .
type ErrorHandler func(error) error

// IngoreError .
func IngoreError(error) error { return nil }

// ErrorAbort .
func ErrorAbort(err error) error { return err }

// ParallelBatch .
func ParallelBatch(
	writers func(i uint64) Writer,
	parallelism,
	size uint64,
	timeout time.Duration,
	errorh ErrorHandler,
) Writer {
	if parallelism <= 0 {
		if size <= 1 {
			return writers(0)
		}
		parallelism = 1
	}
	if size <= 1 {
		writer := &channelWriter{
			dataCh:      make(chan interface{}, parallelism),
			errorCh:     make(chan error, parallelism),
			parallelism: parallelism,
		}
		for i := uint64(0); i < parallelism; i++ {
			go func(w Writer, in *channelWriter) {
				var err error
				defer func() {
					cerr := w.Close()
					if err == nil {
						err = cerr
					}
					writer.errorCh <- err
				}()
				for data := range in.dataCh {
					err = w.Write(data)
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
			}(writers(i), writer)
		}
		return writer
	}
	writer := &channelWriter{
		dataCh:      make(chan interface{}, parallelism*size),
		errorCh:     make(chan error, parallelism),
		parallelism: parallelism,
	}
	for i := uint64(0); i < parallelism; i++ {
		go func(w Writer, in *channelWriter) {
			buf := NewBuffer(w, int(size))
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
				writer.errorCh <- err
			}()
			for {
				select {
				case data, ok := <-in.dataCh:
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
		}(writers(i), writer)
	}
	return writer
}

type channelWriter struct {
	dataCh      chan interface{}
	errorCh     chan error
	parallelism uint64
}

func (w *channelWriter) Write(data interface{}) error {
	w.dataCh <- data
	return nil
}

func (w *channelWriter) WriteN(data ...interface{}) (int, error) {
	for _, item := range data {
		w.dataCh <- item
	}
	return len(data), nil
}

func (w *channelWriter) Close() error {
	close(w.dataCh)
	var errs errorx.Errors
	for i := uint64(0); i < w.parallelism; i++ {
		err := <-w.errorCh
		if err != nil {
			errs = append(errs, err)
		}
	}
	close(w.errorCh)
	return errs.MaybeUnwrap()
}
