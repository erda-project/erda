package server

import (
	"bytes"
	"encoding/json"
	"os"
	"time"

	"github.com/erda-project/erda/pkg/httpclient"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// call this in form of goroutine
func getDCOSTokenAuthPeriodically() {
	client := httpclient.New()
	waitTime := 10 * time.Millisecond

	for {
		select {
		case <-time.After(waitTime):
			token, err := getTokenAuthAndSetEnv(client)

			if err != nil {
				waitTime = 2 * time.Minute
				os.Setenv("AUTH_TOKEN", "")
				logrus.Errorf("get auth token error: %v", err)
				break
			}

			// 每24小时更新一次
			waitTime = 24 * time.Hour

			if len(token) > 0 {
				os.Setenv("AUTH_TOKEN", token)
				logrus.Debugf("get auth token: %s", token)
			} else {
				// err为nil且token为空表示用户没有设置token auth
				os.Unsetenv("AUTH_TOKEN")
				logrus.Debugf("clear auth token")
			}
		}
	}
}

func getTokenAuthAndSetEnv(client *httpclient.HTTPClient) (string, error) {
	dcosAddr := os.Getenv("DCOS_ADDR")
	id := os.Getenv("DCOS_UID")
	password := os.Getenv("DCOS_PASSWORD")

	// uid and password required
	if len(id) == 0 || len(password) == 0 {
		return "", nil
	}
	// dcosAddr is optional, default is internal dcos cluster addr
	if len(dcosAddr) == 0 {
		dcosAddr = "master.mesos"
	}

	logrus.Debugf("id: %v, password: %v, dcosAddr: %v", id, password, dcosAddr)
	var b bytes.Buffer
	type IdAndPassword struct {
		Uid      string `json:"uid"`
		Password string `json:"password"`
	}
	requestIdAndPwd := IdAndPassword{
		Uid:      id,
		Password: password,
	}

	type Token struct {
		AuthToken string `json:"token,omitempty"`
	}
	var token Token

	resp, err := client.Post(dcosAddr).
		Path("/acs/api/v1/auth/login").
		JSONBody(&requestIdAndPwd).
		Header("Content-Type", "application/json").
		Do().
		Body(&b)

	if err != nil {
		return "", err
	}

	if !resp.IsOK() {
		return "", errors.Errorf("get token auth status code: %v, body: %v", resp.StatusCode(), b.String())
	}

	r := bytes.NewReader(b.Bytes())
	if err := json.NewDecoder(r).Decode(&token); err != nil {
		return "", err
	}

	return token.AuthToken, nil
}
