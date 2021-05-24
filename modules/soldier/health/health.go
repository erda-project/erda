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
