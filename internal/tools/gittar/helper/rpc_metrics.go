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

package helper

import (
	"io"
	"net/http"
	"time"

	"github.com/erda-project/erda/internal/tools/gittar/pkg/util/guid"
	"github.com/erda-project/erda/internal/tools/gittar/rpcmetrics"
	"github.com/erda-project/erda/internal/tools/gittar/webcontext"
)

func wrapWithMetrics(c *webcontext.Context, service, version string, reqBody io.ReadCloser, writer io.Writer) (io.ReadCloser, io.Writer, func(error)) {
	return wrapWithMetricsInternal(c, service, version, "advertise", reqBody, writer, nil)
}

func wrapWithMetricsProcess(c *webcontext.Context, service, version string, reqBody io.ReadCloser, writer io.Writer) (io.ReadCloser, io.Writer, func(error)) {
	var cmdCapture *rpcmetrics.LimitedBuffer
	if service == "upload-pack" {
		cmdCapture = rpcmetrics.NewLimitedBuffer(64 * 1024)
		reqBody = rpcmetrics.NewCaptureReadCloser(reqBody, cmdCapture)
	}
	return wrapWithMetricsInternal(c, service, version, "rpc", reqBody, writer, cmdCapture)
}

func wrapWithMetricsInternal(c *webcontext.Context, service, version, phase string, reqBody io.ReadCloser, writer io.Writer, cmdCapture *rpcmetrics.LimitedBuffer) (io.ReadCloser, io.Writer, func(error)) {
	startTime := time.Now()
	userID := ""
	if c.User != nil {
		userID = c.User.Id
	}
	spanID := guid.NewString()

	reqCounter := rpcmetrics.NewCountingReadCloser(reqBody)
	respCounter := rpcmetrics.NewCountingWriter(writer)
	// wrapped
	reqBody = reqCounter
	writer = respCounter

	rpcmetrics.Record(rpcmetrics.Event{
		Timestamp:   startTime,
		Event:       "start",
		ID:          spanID,
		Service:     service,
		Phase:       phase,
		Method:      c.HttpRequest().Method,
		Path:        c.HttpRequest().URL.Path,
		RepoID:      c.Repository.ID,
		RepoPath:    c.Repository.Path,
		OrgID:       c.Repository.OrgId,
		OrgName:     c.Repository.OrgName,
		ProjectID:   c.Repository.ProjectId,
		Project:     c.Repository.ProjectName,
		AppID:       c.Repository.ApplicationId,
		App:         c.Repository.ApplicationName,
		GitProtocol: version,
		RemoteIP:    c.EchoContext.RealIP(),
		UserAgent:   c.HttpRequest().UserAgent(),
		UserID:      userID,
	})

	done := func(err error) {
		status := c.EchoContext.Response().Status
		if status == 0 {
			status = http.StatusOK
		}
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}

		event := rpcmetrics.Event{
			Timestamp:      time.Now(),
			Event:          "end",
			ID:             spanID,
			Service:        service,
			Phase:          phase,
			Method:         c.HttpRequest().Method,
			Path:           c.HttpRequest().URL.Path,
			RepoID:         c.Repository.ID,
			RepoPath:       c.Repository.Path,
			OrgID:          c.Repository.OrgId,
			OrgName:        c.Repository.OrgName,
			ProjectID:      c.Repository.ProjectId,
			Project:        c.Repository.ProjectName,
			AppID:          c.Repository.ApplicationId,
			App:            c.Repository.ApplicationName,
			GitProtocol:    version,
			StartTimestamp: startTime,
			Status:         status,
			DurationMS:     time.Since(startTime).Milliseconds(),
			Error:          errMsg,
		}
		if service == "upload-pack" && cmdCapture != nil {
			event.Cmd, event.CmdParams = rpcmetrics.ParseUploadPackCmd(cmdCapture.Bytes())
		}
		if c.User != nil {
			event.UserID = c.User.Id
		}
		if reqCounter != nil {
			event.ReqBytes = reqCounter.Bytes()
		}
		if respCounter != nil {
			event.RespBytes = respCounter.Bytes()
		}
		event.RemoteIP = c.EchoContext.RealIP()
		event.UserAgent = c.HttpRequest().UserAgent()
		rpcmetrics.Record(event)
	}

	if service == "upload-pack" && cmdCapture != nil {
		go func(id string, buf *rpcmetrics.LimitedBuffer) {
			for i := 0; i < 20; i++ {
				if len(buf.Bytes()) > 0 {
					break
				}
				time.Sleep(50 * time.Millisecond)
			}
			cmd, params := rpcmetrics.ParseUploadPackCmd(buf.Bytes())
			rpcmetrics.UpdateActiveCmd(id, cmd, params)
		}(spanID, cmdCapture)
	}

	return reqBody, writer, done
}
