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
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

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

func (a *Addon) initMySQLAccount(addonIns *dbclient.AddonInstance, addonInsRouting *dbclient.AddonInstanceRouting) error {
	if addonIns.AddonName != "mysql" {
		return nil
	}
	extra, err := a.db.GetByInstanceIDAndField(addonInsRouting.RealInstance, "password")
	if err != nil {
		return err
	}
	if extra == nil {
		return fmt.Errorf("not found extra for instance: %s", addonInsRouting.RealInstance)
	}
	account := dbclient.MySQLAccount{
		Username:          "mysql",
		Password:          extra.Value,
		KMSKey:            addonIns.KmsKey,
		InstanceID:        addonIns.ID,
		RoutingInstanceID: addonInsRouting.ID,
		Creator:           "",
	}
	return a.db.CreateMySQLAccount(&account)
}
