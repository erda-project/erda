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

package middleware

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/steve/websocket"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	// audit template name
	auditCordonNode     = "cordonNode"
	auditUncordonNode   = "uncordonNode"
	auditLabelNode      = "labelNode"
	auditUnLabelNode    = "unLabelNode"
	auditUpdateResource = "updateK8SResource"
	auditCreateResource = "createK8SResource"
	auditDeleteResource = "deleteK8SResource"
	auditKubectlShell   = "kubectlShell"

	// audit template params
	auditClusterName  = "clusterName"
	auditNamespace    = "namespace"
	auditResourceType = "resourceType"
	auditResourceName = "name"
	auditTargetLabel  = "targetLabel"
	auditCommands     = "commands"

	maxAuditLength = 40000
)

type Auditor struct {
	bdl *bundle.Bundle
}

// NewAuditor return a steve Auditor with bundle.
// bdl needs withCoreServices to create audit events.
func NewAuditor(bdl *bundle.Bundle) *Auditor {
	return &Auditor{bdl: bdl}
}

// AuditMiddleWare audit for steve server by bundle.
func (a *Auditor) AuditMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		vars, _ := req.Context().Value(varsKey).(map[string]string)
		var body []byte
		if req.Body != nil {
			body, _ = ioutil.ReadAll(req.Body)
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		writer := &wrapWriter{
			ResponseWriter: resp,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(writer, req)

		if body == nil {
			return
		}
		if writer.statusCode < 200 || writer.statusCode >= 300 {
			return
		}

		clusterName := vars["clusterName"]
		typ := vars["type"]

		namespace := vars["namespace"]
		name := vars["name"]
		isInternal := req.Header.Get(httputil.InternalHeader) != ""
		userID := req.Header.Get(httputil.UserHeader)
		orgID := req.Header.Get(httputil.OrgHeader)
		scopeID, _ := strconv.ParseUint(orgID, 10, 64)
		now := strconv.FormatInt(time.Now().Unix(), 10)

		auditReq := apistructs.AuditCreateRequest{
			Audit: apistructs.Audit{
				UserID:    userID,
				ScopeType: apistructs.OrgScope,
				ScopeID:   scopeID,
				OrgID:     scopeID,
				Result:    "success",
				StartTime: now,
				EndTime:   now,
				ClientIP:  getRealIP(req),
				UserAgent: req.UserAgent(),
			},
		}

		ctx := make(map[string]interface{})

		if vars["kubectl-shell"] == "true" {
			auditReq.TemplateName = auditKubectlShell
			var cmds []string
			for _, cwt := range writer.wc.cmds {
				cmd := fmt.Sprintf("%s: %s", cwt.start.Format(time.RFC3339), cwt.cmd)
				cmds = append(cmds, cmd)
			}
			res := strings.Join(cmds, "\n")
			if len(res) > maxAuditLength {
				res = res[:maxAuditLength] + "..."
			}
			ctx[auditCommands] = res
		} else {
			if typ == "" {
				return
			}
			switch req.Method {
			case http.MethodPatch:
				if isInternal && strutil.Equal(typ, "nodes", true) {
					var rb reqBody
					if err := json.Unmarshal(body, &rb); err != nil {
						logrus.Errorf("failed to unmarshal in steve audit")
						return
					}

					// audit for label/unlabel node
					if rb.Metadata != nil && rb.Metadata["labels"] != nil {
						labels, _ := rb.Metadata["labels"].(map[string]interface{})
						var (
							k string
							v interface{}
						)
						for k, v = range labels {
						} // there can only be one piece of k/v
						if v == nil {
							auditReq.Audit.TemplateName = auditUnLabelNode
							ctx[auditTargetLabel] = k
						} else {
							auditReq.Audit.TemplateName = auditLabelNode
							ctx[auditTargetLabel] = fmt.Sprintf("%s=%s", k, v.(string))
						}
						break
					}

					// audit for cordon/uncordon node
					if rb.Spec != nil && rb.Spec["unschedulable"] != nil {
						v, _ := rb.Spec["unschedulable"].(bool)
						if v {
							auditReq.Audit.TemplateName = auditCordonNode
						} else {
							auditReq.Audit.TemplateName = auditUncordonNode
						}
					}
					break
				}
				fallthrough
			case http.MethodPut:
				auditReq.Audit.TemplateName = auditUpdateResource
			case http.MethodPost:
				auditReq.Audit.TemplateName = auditCreateResource
				var rb reqBody
				if err := json.Unmarshal(body, &rb); err != nil {
					logrus.Errorf("failed to unmarshal in steve audit")
					return
				}
				data := rb.Metadata["name"]
				if n, ok := data.(string); ok && n != "" {
					name = n
				}
				data = rb.Metadata["namespace"]
				if ns, ok := data.(string); ok && namespace != "" {
					namespace = ns
				}
			case http.MethodDelete:
				auditReq.Audit.TemplateName = auditDeleteResource
			default:
				return
			}

			ctx[auditClusterName] = clusterName
			ctx[auditResourceName] = name
			ctx[auditNamespace] = namespace
			ctx[auditResourceType] = typ
			auditReq.Context = ctx
		}
		if err := a.bdl.CreateAuditEvent(&auditReq); err != nil {
			logrus.Errorf("faild to audit in steve audit, %v", err)
		}
	})
}

type reqBody struct {
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Spec     map[string]interface{} `json:"spec,omitempty"`
}

type wrapWriter struct {
	http.ResponseWriter
	statusCode int
	buf        bytes.Buffer
	wc         *wrapConn
}

func (w *wrapWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func (w *wrapWriter) Write(body []byte) (int, error) {
	w.buf.Write(body)
	return w.ResponseWriter.Write(body)
}

func (w *wrapWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("upstream ResponseWriter of type %v does not implement http.Hijacker", reflect.TypeOf(w.ResponseWriter))
	}
	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}

	wc := &wrapConn{
		Conn: conn,
	}
	w.wc = wc
	return wc, rw, nil
}

type wrapConn struct {
	net.Conn
	buf  []byte
	cmds []*cmdWithTimestamp
}

type cmdWithTimestamp struct {
	start time.Time
	cmd   string
}

func (w *wrapConn) Read(p []byte) (n int, err error) {
	n, err = w.Conn.Read(p)
	data := websocket.DecodeFrame(p)
	if len(data) <= 1 {
		return
	}
	decoded, _ := base64.StdEncoding.DecodeString(string(data[1:]))
	if err != nil || len(decoded) == 0 {
		return
	}

	w.buf = append(w.buf, decoded...)
	cmds := strings.Split(string(w.buf), "\n")
	w.buf = nil
	length := len(cmds)
	if decoded[len(decoded)-1] != '\n' {
		w.buf = append(w.buf, cmds[length-1]...)
		length--
	}
	for i := 0; i < length; i++ {
		if len(cmds[i]) == 0 {
			continue
		}
		w.cmds = append(w.cmds, &cmdWithTimestamp{
			start: time.Now(),
			cmd:   cmds[i],
		})
	}
	return
}

func getRealIP(request *http.Request) string {
	ra := request.RemoteAddr
	if ip := request.Header.Get("X-Forwarded-For"); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := request.Header.Get("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}
