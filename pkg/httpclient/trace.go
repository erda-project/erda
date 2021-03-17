package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Tracer interface {
	TraceRequest(*http.Request)
	// if read response body, you need to set it back
	TraceResponse(*http.Response)
}

type DefaultTracer struct {
	w io.Writer
}

func NewDefaultTracer(w io.Writer) *DefaultTracer {
	return &DefaultTracer{w}
}

func (t *DefaultTracer) TraceRequest(req *http.Request) {
	s := fmt.Sprintf("RequestURL: %s\n", req.URL.String())
	io.WriteString(t.w, s)
}

func (t *DefaultTracer) TraceResponse(r *http.Response) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		io.WriteString(t.w, fmt.Sprintf("TraceResponse: read response body fail: %v", err))
		return
	}
	io.WriteString(t.w, fmt.Sprintf("ResponseBody: %s\n", string(body)))
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
}
