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
	"bytes"
	"encoding/hex"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Event struct {
	Timestamp time.Time `json:"ts"`

	Event string `json:"event"`
	ID    string `json:"id,omitempty"`

	Service string `json:"service"`
	Phase   string `json:"phase"`

	Method string `json:"method,omitempty"`
	Path   string `json:"path,omitempty"`

	Cmd       string `json:"cmd,omitempty"`
	CmdParams string `json:"cmd_params,omitempty"`

	RepoID    int64  `json:"repo_id,omitempty"`
	RepoPath  string `json:"repo_path,omitempty"`
	OrgID     int64  `json:"org_id,omitempty"`
	OrgName   string `json:"org,omitempty"`
	ProjectID int64  `json:"project_id,omitempty"`
	Project   string `json:"project,omitempty"`
	AppID     int64  `json:"app_id,omitempty"`
	App       string `json:"app,omitempty"`

	UserID    string `json:"user_id,omitempty"`
	RemoteIP  string `json:"remote_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`

	GitProtocol string `json:"git_protocol,omitempty"`

	StartTimestamp time.Time `json:"start_ts,omitempty"`
	Status         int       `json:"status,omitempty"`
	DurationMS     int64     `json:"duration_ms,omitempty"`
	ReqBytes       int64     `json:"req_bytes,omitempty"`
	RespBytes      int64     `json:"resp_bytes,omitempty"`

	Error string `json:"error,omitempty"`
}

var (
	startedTotal uint64
	endedTotal   uint64
	droppedTotal uint64
)

type activeTracker struct {
	mu     sync.RWMutex
	active map[string]Event
}

var tracker = activeTracker{active: map[string]Event{}}

type Counters struct {
	Started uint64 `json:"started"`
	Ended   uint64 `json:"ended"`
	Dropped uint64 `json:"dropped"`
}

func GetCounters() Counters {
	return Counters{
		Started: atomic.LoadUint64(&startedTotal),
		Ended:   atomic.LoadUint64(&endedTotal),
		Dropped: atomic.LoadUint64(&droppedTotal),
	}
}

func Record(e Event) {
	if !Enabled() {
		return
	}

	switch e.Event {
	case "start":
		atomic.AddUint64(&startedTotal, 1)
		if e.ID != "" {
			tracker.mu.Lock()
			tracker.active[e.ID] = e
			tracker.mu.Unlock()
		}
	case "end":
		atomic.AddUint64(&endedTotal, 1)
		if e.ID != "" {
			tracker.mu.Lock()
			delete(tracker.active, e.ID)
			tracker.mu.Unlock()
		}
	}

	w := global.Load()
	if w != nil {
		w.record(e)
	}
}

func UpdateActiveCmd(id, cmd, params string) {
	if id == "" {
		return
	}
	if cmd == "" && params == "" {
		return
	}
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	e, ok := tracker.active[id]
	if !ok {
		return
	}
	if cmd != "" {
		e.Cmd = cmd
	}
	if params != "" {
		e.CmdParams = params
	}
	tracker.active[id] = e
}

type CountingReadCloser struct {
	rc io.ReadCloser
	n  int64
}

func NewCountingReadCloser(rc io.ReadCloser) *CountingReadCloser {
	if rc == nil {
		return nil
	}
	return &CountingReadCloser{rc: rc}
}

func (c *CountingReadCloser) Read(p []byte) (int, error) {
	n, err := c.rc.Read(p)
	atomic.AddInt64(&c.n, int64(n))
	return n, err
}

func (c *CountingReadCloser) Close() error { return c.rc.Close() }

func (c *CountingReadCloser) Bytes() int64 { return atomic.LoadInt64(&c.n) }

type CountingWriter struct {
	w io.Writer
	n int64
}

func NewCountingWriter(w io.Writer) *CountingWriter {
	if w == nil {
		return nil
	}
	return &CountingWriter{w: w}
}

func (c *CountingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	atomic.AddInt64(&c.n, int64(n))
	return n, err
}

func (c *CountingWriter) Bytes() int64 { return atomic.LoadInt64(&c.n) }

type LimitedBuffer struct {
	buf bytes.Buffer
	max int
}

func NewLimitedBuffer(max int) *LimitedBuffer {
	return &LimitedBuffer{max: max}
}

func (b *LimitedBuffer) Write(p []byte) (int, error) {
	if b.max <= 0 {
		return len(p), nil
	}
	remain := b.max - b.buf.Len()
	if remain > 0 {
		if len(p) <= remain {
			_, _ = b.buf.Write(p)
		} else {
			_, _ = b.buf.Write(p[:remain])
		}
	}
	return len(p), nil
}

func (b *LimitedBuffer) Bytes() []byte {
	return b.buf.Bytes()
}

type teeReadCloser struct {
	r io.Reader
	c io.Closer
}

func (t *teeReadCloser) Read(p []byte) (int, error) { return t.r.Read(p) }
func (t *teeReadCloser) Close() error               { return t.c.Close() }

func NewCaptureReadCloser(rc io.ReadCloser, buf *LimitedBuffer) io.ReadCloser {
	if rc == nil || buf == nil {
		return rc
	}
	return &teeReadCloser{r: io.TeeReader(rc, buf), c: rc}
}

func ParseUploadPackCmd(data []byte) (string, string) {
	cmd := ""
	params := map[string]string{}

	for i := 0; i+4 <= len(data); {
		n, ok := parsePktLen(data[i : i+4])
		if !ok {
			break
		}
		if n == 0 || n == 1 {
			i += 4
			continue
		}
		if n < 4 || i+n > len(data) {
			break
		}
		line := data[i+4 : i+n]
		i += n

		text := strings.TrimSuffix(string(line), "\n")
		if idx := strings.IndexByte(text, 0); idx >= 0 {
			text = text[:idx]
		}
		text = strings.TrimRight(text, "\x00")
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "command=") {
			cmd = strings.TrimPrefix(text, "command=")
			continue
		}
		if strings.HasPrefix(text, "deepen ") {
			params["deepen"] = strings.TrimPrefix(text, "deepen ")
			continue
		}
		if strings.HasPrefix(text, "deepen-since ") {
			params["deepen_since"] = strings.TrimPrefix(text, "deepen-since ")
			continue
		}
		if strings.HasPrefix(text, "deepen-not ") {
			params["deepen_not"] = strings.TrimPrefix(text, "deepen-not ")
			continue
		}
		if strings.HasPrefix(text, "filter ") {
			params["filter"] = strings.TrimPrefix(text, "filter ")
			continue
		}
	}

	if len(params) == 0 {
		return cmd, ""
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	return cmd, strings.Join(parts, " ")
}

func parsePktLen(b []byte) (int, bool) {
	if len(b) != 4 {
		return 0, false
	}
	if bytes.EqualFold(b, []byte("0000")) {
		return 0, true
	}
	decoded := make([]byte, 2)
	if _, err := hex.Decode(decoded, b); err != nil {
		return 0, false
	}
	return int(decoded[0])<<8 | int(decoded[1]), true
}
