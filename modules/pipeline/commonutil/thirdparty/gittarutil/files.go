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
