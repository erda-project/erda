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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

//// GetMySQLAccountList returns a list of MySQL accounts
//func (a *Addon) GetMySQLAccountList(routingInstanceID string) ([]apistructs.MySQLAccount, error) {
//	accounts, err := a.db.GetMySQLAccountListByRoutingInstanceID(routingInstanceID)
//	if err != nil {
//		return nil, err
//	}
//	res := make([]apistructs.MySQLAccount, len(accounts))
//	for i, acc := range accounts {
//		res[i] = apistructs.MySQLAccount{
//			ID:         acc.ID,
//			InstanceID: acc.InstanceRoutingID,
//			Creator:    acc.Creator,
//			Username:   acc.Username,
//			Password:   acc.Password, // TODO: blur
//		}
//	}
//	return res, nil
//}

//func (a *Addon) GenerateMySQLAccount(userID, routingInstanceID string) error {
//	routingInstance, err := a.db.GetInstanceRouting(routingInstanceID)
//	if err != nil {
//		return err
//	}
//	res, err := a.getAddonConfig(routingInstance.RealInstance, nil)
//	if err != nil {
//		return err
//	}
//
//	// TODO: create mysql account in real
//	_ = res
//	// res.Config["MYSQL_HOST"]
//
//	username := "mysql-" + strutil.RandStr(6)
//	password := strutil.RandStr(12)
//
//	kr, err := a.bdl.KMSCreateKey(apistructs.KMSCreateKeyRequest{
//		CreateKeyRequest: kmstypes.CreateKeyRequest{
//			PluginKind: kmstypes.PluginKind_DICE_KMS,
//		},
//	})
//	if err != nil {
//		return err
//	}
//	encryptData, err := a.bdl.KMSEncrypt(apistructs.KMSEncryptRequest{
//		EncryptRequest: kmstypes.EncryptRequest{
//			KeyID:           kr.KeyMetadata.KeyID,
//			PlaintextBase64: base64.StdEncoding.EncodeToString([]byte(password)),
//		},
//	})
//	if err != nil {
//		return err
//	}
//
//	now := time.Now()
//	account := &dbclient.MySQLAccount{
//		CreatedAt:         now,
//		UpdatedAt:         now,
//		Username:          username,
//		Password:          encryptData.CiphertextBase64,
//		KMSKey:            kr.KeyMetadata.KeyID,
//		InstanceID:        routingInstance.RealInstance,
//		InstanceRoutingID: routingInstanceID,
//		Creator:           userID,
//	}
//	return a.db.CreateMySQLAccount(account)
//}

//func (a *Addon) DeleteMySQLAccount(routingInstanceID string, accountID string) error {
//	// TODO: do real delete mysql account
//	account, err := a.db.GetMySQLAccountByID(accountID)
//	if err != nil {
//		return err
//	}
//	account.IsDeleted = true
//	if err := a.db.UpdateMySQLAccount(account); err != nil {
//		return err
//	}
//	return nil
//}

func (a *Addon) toOverrideConfigFromMySQLAccount(config map[string]interface{}, mySQLAccountID string) error {
	account, err := a.db.GetMySQLAccountByID(mySQLAccountID)
	if err != nil {
		return err
	}
	dr, err := a.bdl.KMSDecrypt(apistructs.KMSDecryptRequest{
		DecryptRequest: kmstypes.DecryptRequest{
			KeyID:            account.KMSKey,
			CiphertextBase64: account.Password,
		},
	})
	if err != nil {
		return err
	}
	config["MYSQL_USERNAME"] = account.Username
	config["MYSQL_PASSWORD"] = dr.PlaintextBase64
	return nil
}
