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

package endpoints

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) Logs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	taskIDStr := r.URL.Query().Get("taskID")
	if taskIDStr == "" {
		errstr := "empty taskID arg"
		return httpserver.ErrResp(200, "1", errstr)
	}
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	recordIDstr := r.URL.Query().Get("recordID")
	if recordIDstr == "" {
		errstr := "empty recordID arg"
		return httpserver.ErrResp(200, "1", errstr)
	}
	recordID, err := strconv.ParseUint(recordIDstr, 10, 64)
	if err != nil {
		errstr := fmt.Sprintf("failed to parse recordID: %v", err)
		return httpserver.ErrResp(200, "1", errstr)
	}
	stream := r.URL.Query().Get("stream")
	if stream == "" {
		errstr := "empty stream arg"
		return httpserver.ErrResp(200, "1", errstr)
	}
	start := r.URL.Query().Get("start")
	var startNum, endNum, countNum int64
	if start != "" {
		var err error
		startNum, err = strconv.ParseInt(start, 10, 64)
		if err != nil {
			errstr := fmt.Sprintf("failed to parse 'start' arg: %v", start)
			return httpserver.ErrResp(200, "1", errstr)
		}
	}
	end := r.URL.Query().Get("end")
	if end != "" {
		var err error
		endNum, err = strconv.ParseInt(end, 10, 64)
		if err != nil {
			errstr := fmt.Sprintf("failed to parse 'end' arg: %v", end)
			return httpserver.ErrResp(200, "1", errstr)
		}
	}
	count := r.URL.Query().Get("count")
	if count != "" {
		var err error
		countNum, err = strconv.ParseInt(count, 10, 64)
		if err != nil {
			errstr := fmt.Sprintf("failed to parse 'count' arg: %v", end)
			return httpserver.ErrResp(200, "1", errstr)
		}
	}
	req := apistructs.OpLogsRequest{
		RecordID: recordID,
		TaskID:   taskID,
		Stream:   stream,
		Start:    time.Duration(startNum),
		End:      time.Duration(endNum),
		Count:    countNum,
	}
	logdata, err := e.nodes.Logs(req)
	if err != nil {
		return httpserver.ErrResp(200, "2", err.Error())
	}
	return mkResponse(apistructs.OpLogsResponse{
		Header: apistructs.Header{Success: true},
		Data:   *logdata,
	})
}
