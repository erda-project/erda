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

package cloud_account

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/dbclient"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

type CloudAccount struct {
	db        *dbclient.DBClient
	js        jsonstore.JsonStore
	kmskey    *kmstypes.CreateKeyResponse
	Kmsbundle *bundle.Bundle
}

type VendorAccount struct {
	Org         string
	Vendor      string
	AccessKey   string
	Description string
}

func New(db *dbclient.DBClient, js jsonstore.JsonStore) *CloudAccount {
	return &CloudAccount{db: db, js: js}
}

func (c *CloudAccount) KmsKey() (kmstypes.CreateKeyResponse, error) {
	if c.Kmsbundle == nil {
		c.Kmsbundle = bundle.New(bundle.WithKMS())
	}

	if c.kmskey != nil && c.kmskey.KeyMetadata.KeyID != "" {
		return *c.kmskey, nil
	}
	key := &kmstypes.CreateKeyResponse{}
	if err := c.js.Get(context.Background(), "/dice/ops/kmskey", &key); err != nil {
		if err != jsonstore.NotFoundErr {
			return kmstypes.CreateKeyResponse{}, err
		}
		var err error
		key, err = c.Kmsbundle.KMSCreateKey(apistructs.KMSCreateKeyRequest{
			CreateKeyRequest: kmstypes.CreateKeyRequest{
				PluginKind: kmstypes.PluginKind_DICE_KMS,
			},
		})
		if err != nil {
			return kmstypes.CreateKeyResponse{}, err
		}
		c.kmskey = key
		c.js.Put(context.Background(), "/dice/ops/kmskey", key)
	}
	return *key, nil
}

func (c *CloudAccount) List(org string) ([]VendorAccount, error) {
	akreader := c.db.OrgAKReader()
	aks, err := akreader.ByOrgID(org).Do()
	if err != nil {
		return nil, err
	}
	allaccounts := []VendorAccount{}
	for _, ak := range aks {
		account := VendorAccount{
			Org:         ak.OrgID,
			Vendor:      string(ak.Vendor),
			AccessKey:   ak.AccessKey,
			Description: ak.Description,
		}
		allaccounts = append(allaccounts, account)
	}
	return allaccounts, nil
}

func (c *CloudAccount) Create(org, vendor, ak, secret, desc string) error {
	akreader := c.db.OrgAKReader()

	orgaks, err := akreader.ByOrgID(org).ByVendors(vendor).Do()
	if err != nil {
		return err
	}
	if len(orgaks) != 0 {
		return fmt.Errorf("org: %s, vendor: %s  accessKey exists already", org, vendor)
	}

	kmskey, err := c.KmsKey()
	if err != nil {
		return err
	}
	kmssecret, err := c.Kmsbundle.KMSEncrypt(apistructs.KMSEncryptRequest{
		EncryptRequest: kmstypes.EncryptRequest{
			KeyID:           kmskey.KeyMetadata.KeyID,
			PlaintextBase64: base64.StdEncoding.EncodeToString([]byte(secret)),
		},
	})
	if err != nil {
		return err
	}
	akwriter := c.db.OrgAKWriter()
	orgak := dbclient.OrgAK{
		OrgID:       org,
		Vendor:      dbclient.VendorType(vendor),
		AccessKey:   ak,
		SecretKey:   kmssecret.CiphertextBase64,
		Description: desc,
	}
	if _, err := akwriter.Create(&orgak); err != nil {
		return err
	}
	return nil
}

func (c *CloudAccount) Delete(org, vendor, ak string) error {
	akreader := c.db.OrgAKReader()

	orgaks, err := akreader.ByOrgID(org).ByVendors(vendor).Do()
	if err != nil {
		return err
	}
	for _, ak_ := range orgaks {
		if ak_.AccessKey == ak {
			if err := c.db.OrgAKWriter().Delete(ak_.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
