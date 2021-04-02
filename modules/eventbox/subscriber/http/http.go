package http

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/eventbox/monitor"
	"github.com/erda-project/erda/modules/eventbox/subscriber"
	"github.com/erda-project/erda/modules/eventbox/types"
	"github.com/erda-project/erda/pkg/httpclient"
)

const (
	defaultPublishTimeout = 5 * time.Second
)

// urls
type Dest []string

// label: HTTP-HEADERS
type HTTPHeaders map[string]string

type HTTPSubscriber struct{}

func New() subscriber.Subscriber {
	return &HTTPSubscriber{}
}

func (s *HTTPSubscriber) Publish(dest string, content string, timestamp int64, msg *types.Message) []error {
	monitor.Notify(monitor.MonitorInfo{Tp: monitor.HTTPOutput})

	var d Dest
	dest_ := []byte(dest)
	if err := json.NewDecoder(bytes.NewReader(dest_)).Decode(&d); err != nil {
		return []error{err}
	}
	errs := make(chan error, len(d))
	var wg sync.WaitGroup
	wg.Add(len(d))
	for i := range d {
		destUrl := d[i]
		go func() {
			defer wg.Done()
			if !strings.HasPrefix(destUrl, "http") {
				destUrl = "http://" + destUrl
			}
			parsedUrl, err := url.Parse(destUrl)
			if err != nil {
				errs <- errors.Wrapf(err, "url: %s", destUrl)
				return
			}
			logrus.Debugf("http request url: %s", destUrl)
			buf := bytes.NewBufferString(content)
			opt := []httpclient.OpOption{
				httpclient.WithDnsCache(),
				httpclient.WithDialerKeepAlive(30 * time.Second),
			}
			if parsedUrl.Scheme == "https" {
				opt = []httpclient.OpOption{httpclient.WithHTTPS()}
			}
			var respBody bytes.Buffer
			resp, err := httpclient.New(opt...).Post(parsedUrl.Host).Path(parsedUrl.Path).
				Header("Content-Type", "application/json").RawBody(buf).Do().Body(&respBody)
			if err != nil {
				errs <- errors.Wrapf(err, "url: %s", destUrl)
				return
			}
			if !resp.IsOK() {
				errs <- errors.Errorf("url: %s, response: %d, responseBody: %s", destUrl, resp.StatusCode(), respBody.String())
				logrus.Infof("post content: %v", content)
			} else {
				logrus.Infof("succ HTTP post: %v", parsedUrl)
			}

		}()
	}
	wg.Wait()
	close(errs)
	es := []error{}
	for e := range errs {
		es = append(es, e)
	}
	return es
}

func (s *HTTPSubscriber) Status() interface{} {
	return nil
}

func (s *HTTPSubscriber) Name() string {
	return "HTTP"
}
