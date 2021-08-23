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

package health

import (
	"encoding/json"
	"net/http"
)

const (
	ErdaHealthPath = "/_api/health"
	ErdaHealthy    = "ok"
)

type ModuleHealth struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErdaHealthDto struct {
	Name    string            `json:"name"`
	Status  string            `json:"status"`
	Modules []ModuleHealth    `json:"modules"`
	Tags    map[string]string `json:"tags"`
}

func GetSoldierHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	res, err := json.Marshal(ErdaHealthDto{
		Name:    "soldier",
		Status:  ErdaHealthy,
		Modules: make([]ModuleHealth, 0),
		Tags:    map[string]string{},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
