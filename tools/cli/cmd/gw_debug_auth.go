// Copyright (c) 2022 Terminus, Inc.
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

package cmd

import (
	"crypto/sha256"
	"crypto/sha512"
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
	ctx.Info("auth: %s, ak: %s, sk: %s, content-type: %s, data: %s, method: %s, uri: %s, with timestampe: %v",
		auth, ak, sk, contentType, data, method, uri, timestamp)
	if ak == "" {
		return errors.New("invalid ak")
	}
	if sk == "" {
		return errors.New("invalid sk")
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
	return errors.New("hmac not implement yet")
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
