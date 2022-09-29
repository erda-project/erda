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

package addon

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/http/httpclient"
)

// TODO
func getMyletHost(writeHost string) string {
	return writeHost + ":33080"
}
func getToken(writeHost, localPassword string) string {
	name := strings.TrimSuffix(strings.Split(writeHost, ".")[0], "-write")
	groupToken := hex.EncodeToString(sha256.New().Sum([]byte("root:" + localPassword + "@" + name)))
	return name + "-myctl:0@" + groupToken
}

func createUserDB(username, password, dbname, writeHost, clusterKey string) error {
	q := url.Values{}
	q.Set("username", username)
	q.Set("password", password)
	q.Set("dbname", dbname)
	// q.Set("charset", "")
	// q.Set("collation", "")

	m := make(map[string]interface{}, 2)

	c := httpclient.New(httpclient.WithClusterDialer(clusterKey))
	res, err := c.Post(getMyletHost(writeHost)).
		Path("/api/addons/mylet/user-db").
		Header("Token", getToken(writeHost, password)).
		FormBody(q).
		Do().
		JSON(&m)
	if err != nil {
		return err
	}
	if !res.IsOK() {
		return errors.Errorf("create user&db status code: %d", res.StatusCode())
	}
	if e := m["Error"]; e != nil {
		return errors.Errorf("create user&db return error: %v", e)
	}

	return nil
}

func runSQL(username, password, dbname, initSQL, writeHost, clusterKey string) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fw, err := w.CreateFormField("username")
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, strings.NewReader(username))
	if err != nil {
		return err
	}

	fw, err = w.CreateFormField("password")
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, strings.NewReader(password))
	if err != nil {
		return err
	}

	fw, err = w.CreateFormField("dbname")
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, strings.NewReader(dbname))
	if err != nil {
		return err
	}

	fw, err = w.CreateFormFile("file", filepath.Base(initSQL))
	if err != nil {
		return nil
	}
	f, err := os.Open(initSQL)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, f)
	if err != nil {
		return err
	}

	m := make(map[string]interface{}, 2)

	c := httpclient.New(httpclient.WithClusterDialer(clusterKey))
	res, err := c.Post(getMyletHost(writeHost)).
		Path("/api/addons/mylet/run-sql").
		Header("Token", getToken(writeHost, password)).
		Header("Content-Type", w.FormDataContentType()).
		RawBody(&b).
		Do().
		JSON(&m)
	if err != nil {
		return err
	}
	if !res.IsOK() {
		return errors.Errorf("run sql status code: %d", res.StatusCode())
	}
	if e := m["Error"]; e != nil {
		return errors.Errorf("run sql return error: %v", e)
	}

	return nil
}
