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

package validator

import (
	"math"
	"net/http"
	"reflect"
	"testing"

	"github.com/erda-project/erda/pkg/secret"
)

const (
	mockAuthString                 = "X-Erda-Ak=IQ9E2Buhd8z2h7njPaxeGxq8&X-Erda-Signature=8b6e479c84d975023d6a554e1f252e7cb1efe1b2&X-Erda-Sign-Algorithm=hmac-sha1&X-Erda-Sign-Timestamp=1621123200"
	mockAuthStringWithoutTimestamp = "X-Erda-Ak=IQ9E2Buhd8z2h7njPaxeGxq8&X-Erda-Signature=cb63fc58c286c72827bee6e781f4c3f4c8792347&X-Erda-Sign-Algorithm=hmac-sha1"
	mockAccessKeyID                = "IQ9E2Buhd8z2h7njPaxeGxq8"
	mockSecretKey                  = "0O2Hn0TrTrRwrds1q0un0p9AvX4JB8V6"
)

var mockAkSkPair = secret.AkSkPair{
	AccessKeyID: mockAccessKeyID,
	SecretKey:   mockSecretKey,
}

func mockSignedRequest(inHeader bool, authString string, ops ...func(r *http.Request)) *http.Request {
	r, _ := http.NewRequest(http.MethodGet, "https://example.com/users?page=1&pageNum=10", nil)
	r.Header.Set("x-erda-sdk", "true")
	r.Header.Set("x-erda-version", "0.1.0")
	r.Header.Set("content-type", "application/json")

	if inHeader {
		r.Header.Set("Authorization", authString)
	} else {
		r.URL.RawQuery += "&" + authString
	}

	for _, op := range ops {
		op(r)
	}

	return r
}

func TestHMACValidator_Verify(t *testing.T) {
	type fields struct {
		validator Validator
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Result
	}{
		{
			"in header, with timestamp",
			fields{validator: NewHMACValidator(mockAkSkPair, WithMaxExpireInterval(math.MaxInt64))},
			args{r: mockSignedRequest(true, mockAuthString)},
			Result{
				Ok:      true,
				Message: "",
			},
		},
		{
			"in header, without timestamp",
			fields{validator: NewHMACValidator(mockAkSkPair)},
			args{r: mockSignedRequest(true, mockAuthStringWithoutTimestamp)},
			Result{
				Ok:      true,
				Message: "",
			},
		},
		{
			"in header, without timestamp, with broken req",
			fields{validator: NewHMACValidator(mockAkSkPair)},
			args{r: mockSignedRequest(true, mockAuthStringWithoutTimestamp, func(r *http.Request) {
				r.Header.Set("x-erda-unexpected", "xxx")
			})},
			Result{
				Ok:      false,
				Message: "verify signature failed. got signature: 8b6a33c855fee5a5fc91fb178cab91ba1c59e374",
			},
		},
		{
			"in query string, with timestamp",
			fields{validator: NewHMACValidator(mockAkSkPair, WithMaxExpireInterval(math.MaxInt64))},
			args{r: mockSignedRequest(false, mockAuthString)},
			Result{
				Ok:      true,
				Message: "",
			},
		},
		{
			"in query string, without timestamp",
			fields{validator: NewHMACValidator(mockAkSkPair)},
			args{r: mockSignedRequest(false, mockAuthStringWithoutTimestamp)},
			Result{
				Ok:      true,
				Message: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hv := tt.fields.validator
			if got := hv.Verify(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}
