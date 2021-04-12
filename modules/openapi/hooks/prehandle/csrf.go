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

package prehandle

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/openapi/conf"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	csrfTokenKey   = "OPENAPI-CSRF-TOKEN"
	csrfEncryptKey = []byte("OPENAPI!OPENAPI!OPENAPI!OPENAPI!") // 32 bytes
	csrfExpire     = 12 * time.Hour
	csrfRefresh    = 6 * time.Hour
)

func CSRFToken(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	if req.Header.Get("Referer") == "" {
		return nil
	}
	r, err := url.Parse(req.Header.Get("Referer"))
	if err == nil {
		if strutil.HasSuffixes(r.Host, conf.CSRFWhiteList()...) {
			return nil
		}
	}
	csrftokenCookie, err := req.Cookie(csrfTokenKey)
	if err != nil && (req.Method == "GET" || req.Method == "HEAD" || req.Method == "OPTIONS" || req.Method == "TRACE") {
		// not exist csrftoken in cookie, set it
		if err := setToken(rw, req); err != nil {
			http.Error(rw, fmt.Sprintf("err set csrf token:%v", err), http.StatusInternalServerError)
			return err
		}
		// if not exist csrftoken in cookie, then make it valid
		return nil
	} else if err != nil {
		if err := setToken(rw, req); err != nil {
			http.Error(rw, fmt.Sprintf("err set csrf token:%v", err), http.StatusInternalServerError)
			return err
		}
		err := fmt.Errorf("empty csrf token")
		http.Error(rw, err.Error(), http.StatusForbidden)
		return err
	} else {
		t, err := validateCSRFToken(csrftokenCookie.Value)
		if err != nil {
			if err := setToken(rw, req); err != nil {
				http.Error(rw, fmt.Sprintf("err set csrf token:%v", err), http.StatusInternalServerError)
				return err
			}
		} else {
			now := time.Now()
			if now.After(t.Add(csrfRefresh * time.Hour)) {
				if err := setToken(rw, req); err != nil {
					http.Error(rw, fmt.Sprintf("err set csrf token:%v", err), http.StatusInternalServerError)
					return err
				}
			}
		}
	}
	if req.Method == "GET" || req.Method == "HEAD" || req.Method == "OPTIONS" || req.Method == "TRACE" {
		return nil
	}
	// 非幂等method
	csrftoken := req.Header.Get(csrfTokenKey)
	if csrftoken == "" {
		err := fmt.Errorf("empty csrf token")
		http.Error(rw, err.Error(), http.StatusForbidden)
		return err
	}
	if _, err := validateCSRFToken(csrftoken); err != nil {
		logrus.Warnf("bad csrf token: %v", err)
		http.Error(rw, "bad csrf token", http.StatusForbidden)
		return err
	}
	return nil
}

func setToken(rw http.ResponseWriter, req *http.Request) error {
	token, err := generateCSRFToken()
	if err != nil {
		return err
	}
	reqDomain, err := conf.GetDomain(req.Host, conf.CSRFCookieDomain())
	if err != nil {
		return err
	}
	http.SetCookie(rw, &http.Cookie{
		Name:    csrfTokenKey,
		Value:   token,
		Path:    "/",
		Expires: time.Now().Add(csrfExpire),
		Domain:  reqDomain,
		Secure:  strutil.Contains(conf.DiceProtocol(), "https"),
	})
	return nil
}

func generateCSRFToken() (string, error) {
	buf := make([]byte, 10)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	content := append(buf, []byte(strconv.FormatInt(time.Now().Add(csrfExpire).Unix(), 10))...)
	ciphertext, err := encrypt(content, csrfEncryptKey)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(ciphertext), nil
}

func validateCSRFToken(token string) (*time.Time, error) {
	ciphertext, err := hex.DecodeString(token)
	if err != nil {
		return nil, err
	}
	decryptToken, err := decrypt(ciphertext, csrfEncryptKey)
	if err != nil {
		return nil, err
	}
	timestamp, err := strconv.ParseInt(string(decryptToken[10:]), 10, 64)
	if err != nil {
		return nil, err
	}
	t := time.Unix(timestamp, 0)
	now := time.Now()
	if t.After(now) && t.Before(now.Add(csrfExpire)) {
		return &t, nil
	}
	return nil, fmt.Errorf("illegal csrf token")
}

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
