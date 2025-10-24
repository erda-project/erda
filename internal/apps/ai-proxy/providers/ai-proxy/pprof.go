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

package ai_proxy

import (
	"errors"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

func init() {
	initPprof()
}

func initPprof() {
	runtime.SetMutexProfileFraction(100)
	runtime.SetBlockProfileRate(1)

	go func() {
		server := &http.Server{
			Addr:              ":6060",
			ReadTimeout:       15 * time.Second,
			ReadHeaderTimeout: 15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			Handler:           nil,
		}
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("pprof server ListenAndServe error: %v", err)
		}
	}()
}
