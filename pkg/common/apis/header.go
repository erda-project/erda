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

package apis

import (
	"context"
	"net/http"
	"strconv"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/common/pb"
)

const (
	headerInternalClient = "internal-client"
)

var langKeys = []string{"lang", "accept-language"}

// Language .
func Language(ctx context.Context) i18n.LanguageCodes {
	header := transport.ContextHeader(ctx)
	if header != nil {
		for _, key := range langKeys {
			vals := header.Get(key)
			for _, v := range vals {
				if len(v) > 0 {
					langs, _ := i18n.ParseLanguageCode(v)
					return langs
				}
			}
		}
	}
	return nil
}

// HTTPLanguage .
func HTTPLanguage(r *http.Request) i18n.LanguageCodes {
	lang := r.Header.Get("Lang")
	if len(lang) <= 0 {
		lang = r.Header.Get("Accept-Language")
	}
	langs, _ := i18n.ParseLanguageCode(lang)
	return langs
}

// GetLang .
func GetLang(ctx context.Context) string {
	return GetHeader(ctx, "lang")
}

// GetOrgID .
func GetOrgID(ctx context.Context) string {
	return GetHeader(ctx, "org-id")
}

// GetIntOrgID .
func GetIntOrgID(ctx context.Context) (int64, error) {
	return strconv.ParseInt(GetOrgID(ctx), 10, 64)
}

// GetUserID .
func GetUserID(ctx context.Context) string {
	return GetHeader(ctx, "user-id")
}

// GetIntUserID .
func GetIntUserID(ctx context.Context) (int64, error) {
	return strconv.ParseInt(GetUserID(ctx), 10, 64)
}

// GetHeader
func GetHeader(ctx context.Context, key string) string {
	header := transport.ContextHeader(ctx)
	if header != nil {
		for _, v := range header.Get(key) {
			if len(v) > 0 {
				return v
			}
		}
	}
	return ""
}

func GetInternalClient(ctx context.Context) string {
	return GetHeader(ctx, headerInternalClient)
}

func IsInternalClient(ctx context.Context) bool {
	return GetInternalClient(ctx) != ""
}

// GetIdentityInfo get User-ID and Internal-Client from header.
// return nil if no identity info found.
func GetIdentityInfo(ctx context.Context) *pb.IdentityInfo {
	// try to get User-ID
	userID := GetUserID(ctx)
	internalClient := GetInternalClient(ctx)
	if userID == "" && internalClient == "" {
		return nil
	}
	return &pb.IdentityInfo{UserID: userID, InternalClient: internalClient}
}
