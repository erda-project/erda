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

package cmd

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/tools/cli/command"
)

var now = strconv.FormatInt(time.Now().UnixMilli()/1000, 10)

var GwDebugAuth = command.Command{
	ParentName: "Gw",
	Name:       "debug-auth",
	ShortHelp:  "generation auth command lines",
	LongHelp:   "generation auth command lines",
	Example:    `erda-cli gw debug-auth -X POST --uri 'https://sample.io/api?name=chen&age=28' --content-type application/json --data '{"org": "erda"}' --auth-type sign --ak xxx --sk yyy --timestamp `,
	Flags: []command.Flag{
		command.StringFlag{
			Name:         "auth-type",
			Doc:          Doc("auth type, {'sign', 'hmac'}", "", Optional),
			DefaultValue: "sign",
		},
		command.StringFlag{
			Name: "ak",
			Doc:  Doc("the access key id", "", Required),
		},
		command.StringFlag{
			Name: "sk",
			Doc:  Doc("the secret key id", "", Required),
		},
		command.StringFlag{
			Name:         "content-type",
			Doc:          Doc(`the Content-Type, {'', 'x-www-form-urlencoded', 'application/json}`, "", Optional),
			DefaultValue: "application/json",
		},
		command.StringFlag{
			Name: "data",
			Doc:  Doc("request raw body", "", Optional),
		},
		command.StringFlag{
			Name:         "method",
			Short:        "X",
			Doc:          Doc("http method", "", Optional),
			DefaultValue: http.MethodGet,
		},
		command.StringFlag{
			Name: "uri",
			Doc:  Doc("uri", "", Required),
		},
		command.BoolFlag{
			Name: "timestamp",
			Doc:  Doc("with api timestamp", "", Optional),
		},
	},
	Run: RunGwDebugAuth,
}

func RunGwDebugAuth(ctx *command.Context, auth, ak, sk, contentType, data, method, uri string, timestamp bool) error {
	ctx.Info("auth: %s, ak: %s, sk: %s, content-type: %s, data: %s, method: %s, uri: %s, with timestamp: %v",
		auth, ak, sk, contentType, data, method, uri, timestamp)
	if ak == "" {
		return errors.New("invalid ak (AccessKeyId)")
	}
	if sk == "" {
		return errors.New("invalid sk (AccessKeySecret)")
	}

	u, err := url.Parse(uri)
	if err != nil {
		ctx.Info("failed to parse uri %s", uri)
		return errors.Wrap(err, "invalid uri")
	}

	switch auth {
	case "", "sign", "sign-auth", "signAuth":
		return runGwDebugSignAuth(ctx, ak, sk, contentType, data, method, u, timestamp)
	case "hmac", "hmac-auth", "hmacAuth":
		return runGwDebugHmacAuth(ctx, ak, sk, contentType, data, method, u, timestamp)
	default:
		return errors.Errorf("invalid auth type %s, it must be one of sign or hmac", auth)
	}
}

func runGwDebugSignAuth(ctx *command.Context, ak, sk, contentType, data, method string, uri *url.URL, timestamp bool) error {
	logrus.Infof("content-type: %s", contentType)
	switch contentType {
	case "":
		return signAuthOnUrl(ctx, ak, sk, method, uri, timestamp)
	case "x-www-form-urlencoded":
		return signAuthOnFormUrlencoded(ctx, ak, sk, contentType, data, method, uri, timestamp)
	case "application/json":
		return signAuthOnJson(ctx, ak, sk, contentType, data, method, uri, timestamp)
	default:
		return errors.Errorf("invalid content-type %s, it must be one of x-www-form-urlencoded or application/json", contentType)
	}
}

func runGwDebugHmacAuth(ctx *command.Context, ak, sk, contentType, data, method string, uri *url.URL, timestamp bool) error {
	if len(data) == 0 {
		return hmacAuthWithoutData(ctx, ak, sk, contentType, method, uri)
	}
	return hmacAuthWithData(ctx, ak, sk, contentType, data, method, uri)
}

func signAuthOnUrl(ctx *command.Context, ak, sk, method string, u *url.URL, timestamp bool) error {
	values := u.Query()
	values.Set("appKey", ak)
	if timestamp {
		values.Set("apiTimestamp", now)
	}
	encode := values.Encode()
	logrus.Infof("encoded queries: %s\n", encode)
	signingString := encode + sk
	logrus.Infof("siging string: %s\n", signingString)
	signature := Sha512(signingString)
	logrus.Infof("signature: %s\n", signature)
	values.Add("sign", signature)
	u.RawQuery = values.Encode()
	logrus.Infoln("output curl command")
	fmt.Printf("curl -v -X %s '%s' \n\t",
		strings.ToUpper(method), u.String())
	return nil
}

func signAuthOnFormUrlencoded(ctx *command.Context, ak, sk, contentType, data, method string, u *url.URL, timestamp bool) error {
	values, err := url.ParseQuery(data)
	if err != nil {
		return errors.Wrap(err, "invalid form urlencoded body")
	}
	values.Set("appKey", ak)
	if timestamp {
		values.Set("apiTimestamp", now)
	}
	encode := values.Encode()
	logrus.Infof("encoded queries: %s\n", encode)
	signingString := encode + sk
	logrus.Infof("siging string: %s\n", signingString)
	signature := Sha512(signingString)
	logrus.Infof("signature: %s\n", signature)
	values.Add("sign", signature)
	logrus.Infoln("output curl command")
	fmt.Printf("curl -v -X %s '%s' \\\n\t--header 'Content-Type: %s' \\\n\t--data '%s' \n\t",
		strings.ToUpper(method), u.String(), contentType, strconv.Quote(values.Encode()))
	return nil
}

func signAuthOnJson(ctx *command.Context, ak, sk, contentType, data, method string, u *url.URL, timestamp bool) error {
	var values = make(url.Values)
	if timestamp {
		values.Set("apiTimestamp", now)
	}
	values.Set("appKey", ak)
	values.Set("data", data)
	encode := values.Encode()
	encode, err := url.QueryUnescape(encode)
	if err != nil {
		return err
	}
	logrus.Infof("encoded queries: %s\n", encode)
	signingString := encode + sk
	logrus.Infof("siging string: %s\n", signingString)
	signature := Sha512(signingString)
	logrus.Infof("signature: %s\n", signature)
	var m = make(map[string]string)
	if timestamp {
		m["apiTimestamp"] = now
	}
	m["appKey"] = ak
	m["data"] = data
	m["sign"] = signature
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}
	logrus.Infoln("output curl command")
	fmt.Printf("curl -v -X %s '%s' \\\n\t--header 'Content-Type: %s' \\\n\t--data '%s' \n\t",
		strings.ToUpper(method), u.String(), contentType, string(body))
	return nil
}

func hmacAuthWithData(ctx *command.Context, ak, sk, contentType string, data, method string, u *url.URL) error {
	logrus.Infof("hamc-auth with raw data: %s", data)
	digest, err := HmacGenDigest([]byte(data))
	if err != nil {
		return err
	}
	logrus.Infof("digest: %s", digest)
	date, err := HmacGenDate()
	if err != nil {
		return err
	}
	logrus.Infof("date: %s", date)
	signingString := HmacGenSigningString(date, method, u.RequestURI(), digest)
	logrus.Infof("signing string: %s", signingString)
	signature, err := HmacGenSignature(signingString, sk)
	if err != nil {
		return err
	}
	logrus.Infof("signature: %s", signature)
	authorization := HmacGenAuthorization(ak, signature, digest)
	logrus.Infof("authorization: %s", authorization)
	logrus.Infoln("output curl command")
	fmt.Printf(`curl -v -X %s '%s' \
	-H 'Content-Type: %s' \
	-H 'date: %s' \
	-H 'digest: %s' \
	-H 'Authorization: %s' \
	--data '%s'
`, strings.ToUpper(method), u.String(),
		contentType,
		date,
		digest,
		authorization,
		data,
	)
	return nil
}

func hmacAuthWithoutData(ctx *command.Context, ak, sk, contentType, method string, u *url.URL) error {
	logrus.Infoln("hmac-auth without data")
	date, err := HmacGenDate()
	if err != nil {
		return err
	}
	logrus.Infof("date: %s", date)
	signingString := HmacGenSigningString(date, method, u.RequestURI(), "")
	logrus.Infof("signing string: %s", signingString)
	signature, err := HmacGenSignature(signingString, sk)
	if err != nil {
		return err
	}
	logrus.Infof("signature: %s", signature)
	authorization := HmacGenAuthorization(ak, signature, "")
	logrus.Infof("authorization: %s", authorization)
	logrus.Infoln("output curl command")
	fmt.Printf(`curl -v -X %s '%s' \
	-H 'Content-Type: %s' \
	-H 'date: %s' \
	-H 'Authorization: %s'
`, strings.ToUpper(method), u.String(),
		contentType,
		date,
		authorization,
	)
	return nil
}

func Sha512(s string) string {
	h := sha512.New()
	h.Write([]byte(s))
	data := h.Sum(nil)
	return hex.EncodeToString(data)
}

func GenSHA256HashCode(stringMessage string) string {
	message := []byte(stringMessage) //字符串转化字节数组
	//创建一个基于SHA256算法的hash.Hash接口的对象
	hash := sha256.New() //sha-256加密
	//hash := sha512.New() //SHA-512加密
	//输入数据
	hash.Write(message)
	//计算哈希值
	bytes := hash.Sum(nil)
	//将字符串编码为16进制格式,返回字符串
	hashCode := hex.EncodeToString(bytes)
	//返回哈希值
	return hashCode
}

// HmacGenDigest implements the digest algorithm for raw bodies.
// Note: When computing the signature, encode it from binary to string using the Base64 method.
func HmacGenDigest(data []byte) (string, error) {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return "", err
	}
	return "SHA-256=" + base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// HmacGenDate outputs the current time in the format required by the erda gateway hmac-auth policy.
func HmacGenDate() (string, error) {
	location, err := time.LoadLocation("Etc/GMT+0")
	if err != nil {
		return "", err
	}
	return time.Now().In(location).Format(time.RFC1123), nil
}

// HmacGenSigningString generates the string to be signed
func HmacGenSigningString(date, method, requestLine, digest string) string {
	if digest == "" {
		return fmt.Sprintf("date: %s\n%s %s HTTP/1.1", date, method, requestLine)
	}
	return fmt.Sprintf("date: %s\n%s %s HTTP/1.1\ndigest: %s", date, method, requestLine, digest)
}

// HmacGenSignature generates a signature based on the given date, method, reqeustLine, digest and sk.
func HmacGenSignature(signingString, sk string) (string, error) {
	h := hmac.New(sha256.New, []byte(sk))
	if _, err := h.Write([]byte(signingString)); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// HmacGenAuthorization constructs the value of the request header Authorization
func HmacGenAuthorization(ak, signature, digest string) string {
	if digest == "" {
		return fmt.Sprintf(`Authorization: hmac appkey="%s", algorithm="hmac-sha256", headers="date request-line", signature="%s"`, ak, signature)
	}
	return fmt.Sprintf(`Authorization: hmac appkey="%s", algorithm="hmac-sha256", headers="date request-line digest", signature="%s"`, ak, signature)
}
