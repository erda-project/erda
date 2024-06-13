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

package file

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

var initKmsOnce sync.Once

func (p *provider) applyKmsCmk() (err error) {
	initKmsOnce.Do(func() {
		// check etcd key
		getResp, err := p.etcdClient.Get(context.Background(), p.Cfg.Kms.CmkEtcdKey)
		if err != nil {
			err = fmt.Errorf("failed to get kms files cmk, err: %v", err)
			return
		}
		// not found, apply kms cmk
		if getResp.Count == 0 {
			p.invokeKmsToApplyCmk()
			return
		}
		if getResp.Count > 1 {
			panic(fmt.Errorf("should only have one kv, but got: %d", getResp.Count))
		}
		kv := getResp.Kvs[0]
		p.kmsKey = string(kv.Value)
		if len(p.kmsKey) == 0 {
			p.invokeKmsToApplyCmk()
			return
		} else {
			_, err = p.bdl.KMSRotateKeyVersion(apistructs.KMSRotateKeyVersionRequest{
				RotateKeyVersionRequest: kmstypes.RotateKeyVersionRequest{
					KeyID: p.kmsKey,
				},
			})
			if err != nil {
				panic(fmt.Errorf("files kms cmk rotate key version failed, keyID: %s, err: %v", p.kmsKey, err))
			}
		}
		_, err = p.bdl.KMSDescribeKey(apistructs.KMSDescribeKeyRequest{
			DescribeKeyRequest: kmstypes.DescribeKeyRequest{
				KeyID: p.kmsKey,
			},
		})
		if err != nil {
			logrus.Errorf("files kms cmk describe failed, keyID: %s, err: %v", p.kmsKey, err)
			p.invokeKmsToApplyCmk()
			return
		}
		logrus.Infof("files kms cmk: %s", p.kmsKey)
		return
	})
	return
}

func (p *provider) invokeKmsToApplyCmk() {
	kmsResp, err := p.bdl.KMSCreateKey(apistructs.KMSCreateKeyRequest{
		CreateKeyRequest: kmstypes.CreateKeyRequest{
			PluginKind:            kmstypes.PluginKind_DICE_KMS,
			CustomerMasterKeySpec: kmstypes.CustomerMasterKeySpec_SYMMETRIC_DEFAULT,
			KeyUsage:              kmstypes.KeyUsage_ENCRYPT_DECRYPT,
			Description:           "files kms key",
		},
	})
	if err != nil {
		panic(err)
	}
	keyID := kmsResp.KeyMetadata.KeyID
	defer func() {
		p.kmsKey = keyID
		logrus.Infof("files kms cmk: %s", p.kmsKey)
	}()
	_, err = p.etcdClient.Txn(context.Background()).
		If(clientv3.Compare(clientv3.Version(p.Cfg.Kms.CmkEtcdKey), "=", 0)).
		Then(clientv3.OpPut(p.Cfg.Kms.CmkEtcdKey, keyID)).
		Commit()
	if err != nil {
		panic(err)
	}
	return
}

func (p *provider) GetKMSKey() string {
	return p.kmsKey
}
