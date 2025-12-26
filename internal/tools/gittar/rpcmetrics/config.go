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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Enabled       bool
	Path          string
	BufferSize    int
	FlushInterval time.Duration
}

var (
	enabled    int32
	global     atomic.Pointer[writer]
	globalOnce sync.Once
)

func Enabled() bool { return atomic.LoadInt32(&enabled) == 1 }

func FileWriterEnabled() bool { return global.Load() != nil }

func GetFilePath(t time.Time) string {
	w := global.Load()
	if w == nil {
		return ""
	}
	return w.currentFilePath(t)
}

func Init(cfg Config) error {
	var err error
	globalOnce.Do(func() {
		if !cfg.Enabled {
			atomic.StoreInt32(&enabled, 0)
			global.Store(nil)
			return
		}
		atomic.StoreInt32(&enabled, 1)

		if cfg.BufferSize <= 0 {
			cfg.BufferSize = 1024
		}
		if cfg.FlushInterval <= 0 {
			cfg.FlushInterval = 1 * time.Second
		}
		if cfg.Path == "" {
			return
		}

		var w *writer
		w, err = newWriter(cfg.Path, cfg.BufferSize, cfg.FlushInterval)
		if err != nil {
			return
		}
		global.Store(w)
	})
	return err
}

type CleanupConfig struct {
	// Dir is the directory that contains daily jsonl files, e.g. /repository/.gittar/rpc-metrics
	Dir string
	// KeepDays keeps files whose date is within last N days (inclusive).
	KeepDays int
	// Interval is how often to run cleanup.
	Interval time.Duration
}

func IsDirPath(path string) bool {
	return !strings.HasSuffix(strings.ToLower(path), ".jsonl")
}

func StartCleanupLoop(cfg CleanupConfig) {
	if cfg.Dir == "" || cfg.KeepDays <= 0 || cfg.Interval <= 0 {
		return
	}

	go func() {
		// run once at startup
		runCleanupOnce(cfg.Dir, cfg.KeepDays, time.Now())

		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()
		for range ticker.C {
			runCleanupOnce(cfg.Dir, cfg.KeepDays, time.Now())
		}
	}()
}

func runCleanupOnce(dir string, keepDays int, now time.Time) {
	deleted, err := CleanupOldDailyFiles(dir, keepDays, now)
	if err != nil {
		logrus.Errorf("failed to cleanup git rpc metrics dir %s: %v", dir, err)
		return
	}
	if deleted > 0 {
		logrus.Infof("git rpc metrics cleanup: deleted=%d dir=%s", deleted, dir)
	}
}

// CleanupOldDailyFiles removes files in `dir` named like `YYYY-MM-DD.jsonl` whose date is older than keepDays.
// It does not remove files that do not match the naming pattern.
func CleanupOldDailyFiles(dir string, keepDays int, now time.Time) (int, error) {
	if dir == "" || keepDays <= 0 {
		return 0, nil
	}

	st, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	if !st.IsDir() {
		return 0, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	loc := now.Location()
	cutoff := dateOnly(now.AddDate(0, 0, -keepDays), loc)

	deleted := 0
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".jsonl") {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		day, err := time.ParseInLocation("2006-01-02", base, loc)
		if err != nil {
			continue
		}
		if day.Before(cutoff) {
			if err := os.Remove(filepath.Join(dir, name)); err != nil {
				return deleted, err
			}
			deleted++
		}
	}
	return deleted, nil
}

func dateOnly(t time.Time, loc *time.Location) time.Time {
	tt := t.In(loc)
	return time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, loc)
}
