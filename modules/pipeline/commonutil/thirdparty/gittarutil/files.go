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

package gittarutil

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httpclientutil"
)

func (r *Repo) FetchFile(ref string, filename string) (b []byte, err error) {
	defer func() {
		err = errors.Wrapf(err, "failed to fetch file from gittar, ref [%s], filename [%s]", ref, filename)
	}()
	var content struct {
		Content string `json:"content"`
	}
	req := httpclient.New().Get(r.GittarAddr, httpclient.RetryOption{}).
		Path("/"+r.Repo+"/blob/"+ref+"/"+filename).
		Param("expand", "false").Param("comment", "false")
	if err = httpclientutil.DoJson(req, &content); err != nil {
		return nil, err
	}
	if len(content.Content) == 0 {
		return nil, errors.New("file's content is empty")
	}
	return []byte(content.Content), nil
}
