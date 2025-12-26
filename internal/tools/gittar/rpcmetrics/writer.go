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

package rpcmetrics

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type writer struct {
	basePath string
	isDir    bool

	ch chan Event

	curMu   sync.Mutex
	curDay  string
	curFile *os.File
	curBW   *bufio.Writer
}

func newWriter(path string, bufferSize int, flushInterval time.Duration) (*writer, error) {
	w := &writer{
		basePath: path,
		isDir:    !strings.HasSuffix(strings.ToLower(path), ".jsonl"),
		ch:       make(chan Event, bufferSize),
	}
	if err := w.ensureWriter(time.Now()); err != nil {
		return nil, err
	}
	go w.loop(flushInterval)
	return w, nil
}

func (w *writer) record(e Event) {
	select {
	case w.ch <- e:
	default:
		atomic.AddUint64(&droppedTotal, 1)
		// Avoid blocking git RPC path; drop if the queue is full.
	}
}

func (w *writer) loop(flushInterval time.Duration) {
	flushTicker := time.NewTicker(flushInterval)
	defer flushTicker.Stop()

	for {
		select {
		case e := <-w.ch:
			if err := w.ensureWriter(e.Timestamp); err != nil {
				logrus.Errorf("failed to switch git rpc metrics file: %v", err)
				continue
			}
			b, err := json.Marshal(e)
			if err != nil {
				logrus.Errorf("failed to marshal git rpc metrics event: %v", err)
				continue
			}
			if _, err := w.write(b); err != nil {
				logrus.Errorf("failed to write git rpc metrics event: %v", err)
				continue
			}
			if err := w.writeByte('\n'); err != nil {
				logrus.Errorf("failed to write git rpc metrics newline: %v", err)
				continue
			}
		case <-flushTicker.C:
			// Ensure we rotate even if there is no traffic for a while.
			if err := w.ensureWriter(time.Now()); err != nil {
				logrus.Errorf("failed to switch git rpc metrics file: %v", err)
				continue
			}
			if err := w.flush(); err != nil {
				logrus.Errorf("failed to flush git rpc metrics file %s: %v", w.currentFilePath(time.Now()), err)
			}
		}
	}
}

func (w *writer) currentFilePath(now time.Time) string {
	if !w.isDir {
		return w.basePath
	}
	return filepath.Join(w.basePath, now.Format("2006-01-02")+".jsonl")
}

func (w *writer) ensureWriter(now time.Time) error {
	w.curMu.Lock()
	defer w.curMu.Unlock()

	if !w.isDir {
		if w.curFile != nil && w.curBW != nil {
			if fileExists(w.basePath) {
				return nil
			}
			_ = w.curBW.Flush()
			_ = w.curFile.Close()
			w.curFile = nil
			w.curBW = nil
		}
		if err := os.MkdirAll(filepath.Dir(w.basePath), 0755); err != nil {
			return err
		}
		f, err := os.OpenFile(w.basePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		w.curFile = f
		w.curBW = bufio.NewWriterSize(f, 64*1024)
		return nil
	}

	day := now.Format("2006-01-02")
	filePath := filepath.Join(w.basePath, day+".jsonl")
	if day == w.curDay && w.curFile != nil && w.curBW != nil {
		if fileExists(filePath) {
			return nil
		}
		_ = w.curBW.Flush()
		_ = w.curFile.Close()
		w.curFile = nil
		w.curBW = nil
	}
	if err := os.MkdirAll(w.basePath, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if w.curBW != nil {
		_ = w.curBW.Flush()
	}
	if w.curFile != nil {
		_ = w.curFile.Close()
	}
	w.curDay = day
	w.curFile = f
	w.curBW = bufio.NewWriterSize(f, 64*1024)
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (w *writer) write(p []byte) (int, error) {
	w.curMu.Lock()
	defer w.curMu.Unlock()
	if w.curBW == nil {
		return 0, io.ErrClosedPipe
	}
	return w.curBW.Write(p)
}

func (w *writer) writeByte(b byte) error {
	w.curMu.Lock()
	defer w.curMu.Unlock()
	if w.curBW == nil {
		return io.ErrClosedPipe
	}
	return w.curBW.WriteByte(b)
}

func (w *writer) flush() error {
	w.curMu.Lock()
	defer w.curMu.Unlock()
	if w.curBW == nil {
		return nil
	}
	return w.curBW.Flush()
}
