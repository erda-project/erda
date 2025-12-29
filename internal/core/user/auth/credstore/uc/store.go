package uc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/erda-project/erda/internal/core/user/auth/applier"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
)

type Config struct {
	Redis      *redis.Client
	CookieName string
	Expire     time.Duration
}

type Store struct {
	cfg *Config
}

func (s *Store) Load(_ context.Context, r *http.Request) (*domain.PersistedCredential, error) {
	session, err := r.Cookie(s.cfg.CookieName)
	if err != nil {
		return nil, err
	}

	token, err := s.cfg.Redis.Get(makeSessionKey(session.Value)).Result()
	if err != nil {
		return nil, err
	}

	return &domain.PersistedCredential{
		Authenticator: &applier.BearerTokenAuth{
			Token: token,
		},
		AccessToken: token,
		SessionID:   session.Value,
	}, nil
}

func (s *Store) Persist(_ context.Context, cred *domain.PersistedCredential) (*domain.PersistedCredential, error) {
	if cred.AccessToken == "" {
		return nil, errors.New("credential token is empty")
	}
	sessionID := genSessionID()
	if _, err := s.cfg.Redis.Set(makeSessionKey(sessionID), cred.AccessToken, s.cfg.Expire).Result(); err != nil {
		return nil, errors.Wrap(err, "failed to store session")
	}
	// TODO: new credential with cookie or session?
	cred.SessionID = sessionID
	return cred, nil
}

func (s *Store) Revoke(_ context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	_, err := s.cfg.Redis.Del(makeSessionKey(sessionID)).Result()
	return err
}

func makeSessionKey(sessionID string) string {
	return fmt.Sprintf("openapi:sessionid:%s", sessionID)
}

func genSessionID() string {
	return uuid.NewV4().String()
}
