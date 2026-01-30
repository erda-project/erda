package facade

import (
	"context"
	"net/http"

	identitypb "github.com/erda-project/erda-proto-go/core/user/identity/pb"
)

type PersistedCredential struct {
	Type        identitypb.TokenSource
	AccessToken string `json:"accessToken"`
	// CookieName is set when Type == Cookie (from CredentialStore). Used to build GetCurrentUserRequest.
	CookieName string
}

type CredentialStore interface {
	Load(ctx context.Context, req *http.Request) (*PersistedCredential, error)
}

type cookieStore struct {
	cookieName string
}

func (c *cookieStore) Load(_ context.Context, req *http.Request) (*PersistedCredential, error) {
	cookie, err := req.Cookie(c.cookieName)
	if err != nil {
		return nil, err
	}
	return &PersistedCredential{
		Type:        identitypb.TokenSource_Cookie,
		AccessToken: cookie.Value,
		CookieName:  c.cookieName,
	}, nil
}

func NewCookieStore(cookieName string) CredentialStore {
	return &cookieStore{
		cookieName: cookieName,
	}
}

// ToGetCurrentUserRequest builds identity GetCurrentUserRequest from credential. Used by facade callers (e.g. userState).
func ToGetCurrentUserRequest(cred *PersistedCredential) *identitypb.GetCurrentUserRequest {
	req := &identitypb.GetCurrentUserRequest{
		AccessToken: cred.AccessToken,
		Source:      cred.Type,
	}
	if cred.CookieName != "" {
		req.CookieName = &cred.CookieName
	}
	return req
}
