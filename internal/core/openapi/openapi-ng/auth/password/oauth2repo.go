package password

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	"gopkg.in/oauth2.v3"
	oauth2model "gopkg.in/oauth2.v3/models"

	oauth2store "github.com/erda-project/erda/pkg/oauth2/clientstore/mysqlclientstore"
)

const openapiClientID = "openapi"

type oauth2ClientStore interface {
	GetByID(id string) (oauth2.ClientInfo, error)
	Create(info oauth2.ClientInfo) error
}

type OAuth2Repo struct {
	clientStore oauth2ClientStore
}

func NewOAuth2Repo() (*OAuth2Repo, error) {
	store, err := oauth2store.NewClientStore()
	if err != nil {
		return nil, fmt.Errorf("init oauth2 client store failed: %w", err)
	}
	return &OAuth2Repo{clientStore: store}, nil
}

// GetOrCreateOpenAPIClient returns the openapi oauth2 client,
// creates it if not exists.
func (r *OAuth2Repo) GetOrCreateOpenAPIClient() (*oauth2store.ClientStoreItem, error) {
	item, err := r.clientStore.GetByID(openapiClientID)
	if err == nil {
		return &oauth2store.ClientStoreItem{
			ID:     item.GetID(),
			Secret: item.GetSecret(),
		}, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// not exists → create
	secret, err := genOAuth2Secret()
	if err != nil {
		return nil, err
	}

	client := &oauth2model.Client{
		ID:     openapiClientID,
		Secret: secret,
	}

	if err := r.clientStore.Create(client); err != nil {
		return nil, err
	}

	return &oauth2store.ClientStoreItem{
		ID:     openapiClientID,
		Secret: secret,
	}, nil
}

func genOAuth2Secret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate oauth2 secret failed: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
