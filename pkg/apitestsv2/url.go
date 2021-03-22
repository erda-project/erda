package apitestsv2

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/erda-project/erda/pkg/httpclientutil"
)

// domain: schema://${prefix} + u
func polishURL(rawurl string, domain string) (string, error) {
	if rawurl == "" {
		return "", fmt.Errorf("empty url")
	}

	// 将 URL 后面的参数去掉
	rawurl = strings.Split(rawurl, "?")[0]

	// 若 url 非 http://, https:// 开头，则补充 domain
	if !httpclientutil.HasSchema(rawurl) {
		rawurl = domain + rawurl
	}

	// 保证协议头存在
	rawurl = httpclientutil.WrapProto(rawurl)

	if _, err := url.Parse(rawurl); err != nil {
		return "", fmt.Errorf("invalid url: %s, err: %v", rawurl, err)
	}

	return rawurl, nil
}
