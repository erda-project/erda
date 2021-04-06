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

package colonyutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// TODO Deprecated util
func WriteJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(v)
	if err != nil {
		logrus.Debugln(err)
	}
	_, err = w.Write(b)
	if err != nil {
		logrus.Debugln(err)
	}
}

func WriteData(w http.ResponseWriter, v interface{}) {
	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    v,
	})
}

func WriteErr(w http.ResponseWriter, code, msg string) {
	WriteJSON(w, map[string]interface{}{
		"success": false,
		"err": map[string]interface{}{
			"code": code,
			"msg":  msg,
			"ctx":  nil,
		},
	})
}

func WriteError(w http.ResponseWriter, code int, err error) {
	c := strconv.Itoa(code)
	msg := fmt.Sprintf("%v", err)

	WriteErr(w, c, msg)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("%s host: %s, protocol: %s, method: %s, header: %s, uri: %s",
			time.Now().String(), r.Host, r.Proto, r.Method, r.Header, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
