/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package spdy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/httpstream"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/third_party/forked/golang/netutil"
	"k8s.io/klog"
)

// SpdyRoundTripper knows how to upgrade an HTTP request to one that supports
// multiplexed streams. After RoundTrip() is invoked, Conn will be set
// and usable. SpdyRoundTripper implements the UpgradeRoundTripper interface.
type SpdyRoundTripper struct {
	//tlsConfig holds the TLS configuration settings to use when connecting
	//to the remote server.
	tlsConfig *tls.Config

	/* TODO according to http://golang.org/pkg/net/http/#RoundTripper, a RoundTripper
	   must be safe for use by multiple concurrent goroutines. If this is absolutely
	   necessary, we could keep a map from http.Request to net.Conn. In practice,
	   a client will create an http.Client, set the transport to a new insteace of
	   SpdyRoundTripper, and use it a single time, so this hopefully won't be an issue.
	*/
	// conn is the underlying network connection to the remote server.
	conn net.Conn

	// Dialer is the dialer used to connect.  Used if non-nil.
	DialerTimeout  time.Duration
	DialerDeadline time.Time
	Dialer         func(ctx context.Context, network, address string) (net.Conn, error)

	// proxier knows which proxy to use given a request, defaults to http.ProxyFromEnvironment
	// Used primarily for mocking the proxy discovery in tests.
	proxier func(req *http.Request) (*url.URL, error)

	// followRedirects indicates if the round tripper should examine responses for redirects and
	// follow them.
	followRedirects bool
	// requireSameHostRedirects restricts redirect following to only follow redirects to the same host
	// as the original request.
	requireSameHostRedirects bool
}

var _ utilnet.TLSClientConfigHolder = &SpdyRoundTripper{}
var _ httpstream.UpgradeRoundTripper = &SpdyRoundTripper{}
var _ utilnet.Dialer = &SpdyRoundTripper{}

// NewRoundTripper creates a new SpdyRoundTripper that will use
// the specified tlsConfig.
func NewRoundTripper(tlsConfig *tls.Config, followRedirects, requireSameHostRedirects bool) httpstream.UpgradeRoundTripper {
	return NewSpdyRoundTripper(tlsConfig, followRedirects, requireSameHostRedirects)
}

// NewSpdyRoundTripper creates a new SpdyRoundTripper that will use
// the specified tlsConfig. This function is mostly meant for unit tests.
func NewSpdyRoundTripper(tlsConfig *tls.Config, followRedirects, requireSameHostRedirects bool) *SpdyRoundTripper {
	return &SpdyRoundTripper{
		tlsConfig:                tlsConfig,
		followRedirects:          followRedirects,
		requireSameHostRedirects: requireSameHostRedirects,
	}
}

// TLSClientConfig implements pkg/util/net.TLSClientConfigHolder for proper TLS checking during
// proxying with a spdy roundtripper.
func (s *SpdyRoundTripper) TLSClientConfig() *tls.Config {
	return s.tlsConfig
}

// Dial implements k8s.io/apimachinery/pkg/util/net.Dialer.
func (s *SpdyRoundTripper) Dial(req *http.Request) (net.Conn, error) {
	conn, err := s.dial(req)
	if err != nil {
		return nil, err
	}

	if err := req.Write(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// dial dials the host specified by req, using TLS if appropriate, optionally
// using a proxy server if one is configured via environment variables.
func (s *SpdyRoundTripper) dial(req *http.Request) (net.Conn, error) {
	proxier := s.proxier
	if proxier == nil {
		proxier = utilnet.NewProxierWithNoProxyCIDR(http.ProxyFromEnvironment)
	}
	proxyURL, err := proxier(req)
	if err != nil {
		return nil, err
	}

	if proxyURL == nil {
		return s.dialWithoutProxy(req.Context(), req.URL)
	}

	// ensure we use a canonical host with proxyReq
	targetHost := netutil.CanonicalAddr(req.URL)

	// proxying logic adapted from http://blog.h6t.eu/post/74098062923/golang-websocket-with-http-proxy-support
	proxyReq := http.Request{
		Method: "CONNECT",
		URL:    &url.URL{},
		Host:   targetHost,
	}

	if pa := s.proxyAuth(proxyURL); pa != "" {
		proxyReq.Header = http.Header{}
		proxyReq.Header.Set("Proxy-Authorization", pa)
	}

	proxyDialConn, err := s.dialWithoutProxy(req.Context(), proxyURL)
	if err != nil {
		return nil, err
	}

	proxyClientConn := httputil.NewProxyClientConn(proxyDialConn, nil)
	_, err = proxyClientConn.Do(&proxyReq)
	if err != nil && err != httputil.ErrPersistEOF {
		return nil, err
	}

	rwc, _ := proxyClientConn.Hijack()

	if req.URL.Scheme != "https" {
		return rwc, nil
	}

	host, _, err := net.SplitHostPort(targetHost)
	if err != nil {
		return nil, err
	}

	tlsConfig := s.tlsConfig
	switch {
	case tlsConfig == nil:
		tlsConfig = &tls.Config{ServerName: host}
	case len(tlsConfig.ServerName) == 0:
		tlsConfig = tlsConfig.Clone()
		tlsConfig.ServerName = host
	}

	tlsConn := tls.Client(rwc, tlsConfig)

	// need to manually call Handshake() so we can call VerifyHostname() below
	if err := tlsConn.Handshake(); err != nil {
		return nil, err
	}

	// Return if we were configured to skip validation
	if tlsConfig.InsecureSkipVerify {
		return tlsConn, nil
	}

	if err := tlsConn.VerifyHostname(tlsConfig.ServerName); err != nil {
		return nil, err
	}

	return tlsConn, nil
}

// dialWithoutProxy dials the host specified by url, using TLS if appropriate.
func (s *SpdyRoundTripper) dialWithoutProxy(ctx context.Context, url *url.URL) (net.Conn, error) {
	dialAddr := netutil.CanonicalAddr(url)

	d := net.Dialer{
		Timeout:  s.DialerTimeout,
		Deadline: s.DialerDeadline,
	}
	if url.Scheme == "http" {
		if s.Dialer == nil {
			return d.DialContext(ctx, "tcp", dialAddr)
		} else {
			return s.Dialer(ctx, "tcp", dialAddr)
		}
	}

	// TODO validate the TLSClientConfig is set up?
	var conn *tls.Conn
	var err error
	if s.Dialer == nil {
		conn, err = dialTLS(context.Background(), d.Timeout, d.Deadline, d.DialContext, "tcp", dialAddr, s.tlsConfig)
	} else {
		conn, err = dialTLS(context.Background(), s.DialerTimeout, s.DialerDeadline, s.Dialer, "tcp", dialAddr, s.tlsConfig)
	}
	if err != nil {
		return nil, err
	}

	// Return if we were configured to skip validation
	if s.tlsConfig != nil && s.tlsConfig.InsecureSkipVerify {
		return conn, nil
	}

	host, _, err := net.SplitHostPort(dialAddr)
	if err != nil {
		return nil, err
	}
	if s.tlsConfig != nil && len(s.tlsConfig.ServerName) > 0 {
		host = s.tlsConfig.ServerName
	}
	err = conn.VerifyHostname(host)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// proxyAuth returns, for a given proxy URL, the value to be used for the Proxy-Authorization header
func (s *SpdyRoundTripper) proxyAuth(proxyURL *url.URL) string {
	if proxyURL == nil || proxyURL.User == nil {
		return ""
	}
	credentials := proxyURL.User.String()
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encodedAuth)
}

// RoundTrip executes the Request and upgrades it. After a successful upgrade,
// clients may call SpdyRoundTripper.Connection() to retrieve the upgraded
// connection.
func (s *SpdyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	header := utilnet.CloneHeader(req.Header)
	header.Add(httpstream.HeaderConnection, httpstream.HeaderUpgrade)
	header.Add(httpstream.HeaderUpgrade, HeaderSpdy31)

	var (
		conn        net.Conn
		rawResponse []byte
		err         error
	)

	if s.followRedirects {
		conn, rawResponse, err = ConnectWithRedirects(req.Method, req.URL, header, req.Body, s, s.requireSameHostRedirects)
	} else {
		clone := utilnet.CloneRequest(req)
		clone.Header = header
		conn, err = s.Dial(clone)
	}
	if err != nil {
		return nil, err
	}

	responseReader := bufio.NewReader(
		io.MultiReader(
			bytes.NewBuffer(rawResponse),
			conn,
		),
	)

	resp, err := http.ReadResponse(responseReader, nil)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, err
	}

	s.conn = conn

	return resp, nil
}

func ConnectWithRedirects(originalMethod string, originalLocation *url.URL, header http.Header, originalBody io.Reader, dialer utilnet.Dialer, requireSameHostRedirects bool) (net.Conn, []byte, error) {

	const (
		maxRedirects    = 9     // Fail on the 10th redirect
		maxResponseSize = 16384 // play it safe to allow the potential for lots of / large headers
	)

	var (
		location         = originalLocation
		method           = originalMethod
		intermediateConn net.Conn
		rawResponse      = bytes.NewBuffer(make([]byte, 0, 256))
		body             = originalBody
	)

	defer func() {
		if intermediateConn != nil {
			intermediateConn.Close()
		}
	}()

redirectLoop:
	for redirects := 0; ; redirects++ {
		if redirects > maxRedirects {
			return nil, nil, fmt.Errorf("too many redirects (%d)", redirects)
		}

		req, err := http.NewRequest(method, location.String(), body)
		if err != nil {
			return nil, nil, err
		}

		req.Header = header

		intermediateConn, err = dialer.Dial(req)
		if err != nil {
			return nil, nil, err
		}

		// Peek at the backend response.
		rawResponse.Reset()
		respReader := bufio.NewReader(io.TeeReader(
			io.LimitReader(intermediateConn, maxResponseSize), // Don't read more than maxResponseSize bytes.
			rawResponse)) // Save the raw response.
		resp, err := http.ReadResponse(respReader, nil)
		if err != nil {
			// Unable to read the backend response; let the client handle it.
			klog.Warningf("Error reading backend response: %v", err)
			break redirectLoop
		}

		switch resp.StatusCode {
		case http.StatusFound:
			// Redirect, continue.
		default:
			// Don't redirect.
			break redirectLoop
		}

		// Redirected requests switch to "GET" according to the HTTP spec:
		// https://www.w3.org/Protocols/rfc2616/rfc2616-sec10.html#sec10.3
		method = "GET"
		// don't send a body when following redirects
		body = nil

		resp.Body.Close() // not used

		// Prepare to follow the redirect.
		redirectStr := resp.Header.Get("Location")
		if redirectStr == "" {
			return nil, nil, fmt.Errorf("%d response missing Location header", resp.StatusCode)
		}
		// We have to parse relative to the current location, NOT originalLocation. For example,
		// if we request http://foo.com/a and get back "http://bar.com/b", the result should be
		// http://bar.com/b. If we then make that request and get back a redirect to "/c", the result
		// should be http://bar.com/c, not http://foo.com/c.
		location, err = location.Parse(redirectStr)
		if err != nil {
			return nil, nil, fmt.Errorf("malformed Location header: %v", err)
		}

		// Only follow redirects to the same host. Otherwise, propagate the redirect response back.
		if requireSameHostRedirects && location.Hostname() != originalLocation.Hostname() {
			return nil, nil, fmt.Errorf("hostname mismatch: expected %s, found %s", originalLocation.Hostname(), location.Hostname())
		}

		// Reset the connection.
		intermediateConn.Close()
		intermediateConn = nil
	}

	connToReturn := intermediateConn
	intermediateConn = nil // Don't close the connection when we return it.
	return connToReturn, rawResponse.Bytes(), nil
}

// NewConnection validates the upgrade response, creating and returning a new
// httpstream.Connection if there were no errors.
func (s *SpdyRoundTripper) NewConnection(resp *http.Response) (httpstream.Connection, error) {
	connectionHeader := strings.ToLower(resp.Header.Get(httpstream.HeaderConnection))
	upgradeHeader := strings.ToLower(resp.Header.Get(httpstream.HeaderUpgrade))
	if (resp.StatusCode != http.StatusSwitchingProtocols) || !strings.Contains(connectionHeader, strings.ToLower(httpstream.HeaderUpgrade)) || !strings.Contains(upgradeHeader, strings.ToLower(HeaderSpdy31)) {
		defer resp.Body.Close()
		responseError := ""
		responseErrorBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			responseError = "unable to read error from server response"
		} else {
			// TODO: I don't belong here, I should be abstracted from this class
			if obj, _, err := statusCodecs.UniversalDecoder().Decode(responseErrorBytes, nil, &metav1.Status{}); err == nil {
				if status, ok := obj.(*metav1.Status); ok {
					return nil, &apierrors.StatusError{ErrStatus: *status}
				}
			}
			responseError = string(responseErrorBytes)
			responseError = strings.TrimSpace(responseError)
		}

		return nil, fmt.Errorf("unable to upgrade connection: %s", responseError)
	}

	return NewClientConnection(s.conn)
}

// statusScheme is private scheme for the decoding here until someone fixes the TODO in NewConnection
var statusScheme = runtime.NewScheme()

// ParameterCodec knows about query parameters used with the meta v1 API spec.
var statusCodecs = serializer.NewCodecFactory(statusScheme)

func init() {
	statusScheme.AddUnversionedTypes(metav1.SchemeGroupVersion,
		&metav1.Status{},
	)
}
