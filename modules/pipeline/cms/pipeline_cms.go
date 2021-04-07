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

package cms

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/encryption"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	CtxKeyPipelineSource = "pipelineSource"
	CtxKeyForceDelete    = "forceDelete"
)

type pipelineCm struct {
	dbClient *dbclient.Client
	rsaCrypt *encryption.RsaCrypt
}

func NewPipelineCms(dbClient *dbclient.Client, rsaCrypt *encryption.RsaCrypt) *pipelineCm {
	var cm pipelineCm
	cm.dbClient = dbClient
	cm.rsaCrypt = rsaCrypt
	return &cm
}

func (c *pipelineCm) IdempotentCreateNS(ctx context.Context, ns string) error {
	pipelineSource, err := getPipelineSourceFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = c.dbClient.IdempotentCreateCmsNs(pipelineSource, ns)
	return err
}

func (c *pipelineCm) IdempotentDeleteNS(ctx context.Context, ns string) error {
	pipelineSource, err := getPipelineSourceFromContext(ctx)
	if err != nil {
		return err
	}
	return c.dbClient.IdempotentDeleteCmsNs(pipelineSource, ns)
}

func (c *pipelineCm) PrefixListNS(ctx context.Context, nsPrefix string) ([]apistructs.PipelineCmsNs, error) {
	pipelineSource, err := getPipelineSourceFromContext(ctx)
	if err != nil {
		return nil, err
	}
	namespaces, err := c.dbClient.PrefixListNs(pipelineSource, nsPrefix)
	if err != nil {
		return nil, err
	}
	var result []apistructs.PipelineCmsNs
	for _, ns := range namespaces {
		result = append(result, apistructs.PipelineCmsNs{
			PipelineSource: ns.PipelineSource,
			NS:             ns.Ns,
			TimeCreated:    ns.TimeCreated,
			TimeUpdated:    ns.TimeUpdated,
		})
	}
	return result, nil
}

func (c *pipelineCm) UpdateConfigs(ctx context.Context, ns string, kvs map[string]apistructs.PipelineCmsConfigValue) error {
	pipelineSource, err := getPipelineSourceFromContext(ctx)
	if err != nil {
		return err
	}
	cmsNs, exist, err := c.dbClient.GetCmsNs(pipelineSource, ns)
	if err != nil {
		return err
	}

	var cmsNsID = cmsNs.ID
	if !exist {
		if err := c.IdempotentCreateNS(ctx, ns); err != nil {
			return err
		}
		// 获取 ns
		newCmsNs, exist, err := c.dbClient.GetCmsNs(pipelineSource, ns)
		if err != nil {
			return err
		}
		if !exist {
			return errors.Errorf("failed to get cms ns after create, pipelineSource: %s, ns: %s", pipelineSource, ns)
		}
		cmsNsID = newCmsNs.ID
	}
	var configs []spec.PipelineCmsConfig
	for k, v := range kvs {
		vv, err := c.encryptValueIfNeeded(v.EncryptInDB, v.Value)
		if err != nil {
			return err
		}
		setDefault(&v)
		if err := validateConfigWhenUpdate(k, v); err != nil {
			return errors.Errorf("config key: %s, err: %v", k, err)
		}
		configs = append(configs, spec.PipelineCmsConfig{
			NsID:    cmsNsID,
			Key:     k,
			Value:   vv,
			Encrypt: &[]bool{v.EncryptInDB}[0],
			Type:    v.Type,
			Extra: spec.PipelineCmsConfigExtra{
				Operations: v.Operations,
				Comment:    v.Comment,
				From:       v.From,
			},
		})
	}
	// tx
	txSession := c.dbClient.NewSession()
	defer txSession.Close()
	if err := txSession.Begin(); err != nil {
		return err
	}
	err = c.dbClient.UpdateCmsNsConfigs(cmsNs, configs, dbclient.WithTxSession(txSession.Session))
	if err != nil {
		if rbErr := txSession.Rollback(); rbErr != nil {
			logrus.Errorf("[alert] failed to rollback tx session when update pipeline cms ns configs failed, pipelineSource: %s, ns: %s, rbErr: %v, err: %v",
				pipelineSource, ns, rbErr, err)
			return err
		}
		return err
	}
	if cmErr := txSession.Commit(); cmErr != nil {
		logrus.Errorf("[alert] failed to commit tx session when update pipeline cms ns configs success, pipelineSource: %s, ns: %s, cmErr: %v, err: %v",
			pipelineSource, ns, cmErr, err)
		return cmErr
	}
	return nil
}

func (c *pipelineCm) DeleteConfigs(ctx context.Context, ns string, keys ...string) error {
	pipelineSource, err := getPipelineSourceFromContext(ctx)
	if err != nil {
		return err
	}
	cmsNs, exist, err := c.dbClient.GetCmsNs(pipelineSource, ns)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}

	// validate keys before delete
	configs, err := c.GetConfigs(ctx, ns, false, func() []apistructs.PipelineCmsConfigKey {
		var getKeys []apistructs.PipelineCmsConfigKey
		for _, key := range keys {
			getKeys = append(getKeys, apistructs.PipelineCmsConfigKey{
				Key:     key,
				Decrypt: false,
			})
		}
		return getKeys
	}()...)
	if err != nil {
		return err
	}
	var cannotDelKeys []string
	forceDel, ok := ctx.Value(CtxKeyForceDelete).(bool)
	if !ok {
		return errors.New("failed to get force delete key")
	}
	if !forceDel {
		for k, v := range configs {
			if !v.Operations.CanDelete {
				cannotDelKeys = append(cannotDelKeys, k)
			}
		}
		if len(cannotDelKeys) > 0 {
			return errors.Errorf("cannot delete keys: %s", strutil.Join(cannotDelKeys, ", ", true))
		}
	}

	return c.dbClient.DeleteCmsNsConfigs(cmsNs, keys)
}

func (c *pipelineCm) GetConfigs(ctx context.Context, ns string, globalDecrypt bool, keys ...apistructs.PipelineCmsConfigKey) (map[string]apistructs.PipelineCmsConfigValue, error) {
	pipelineSource, err := getPipelineSourceFromContext(ctx)
	if err != nil {
		return nil, err
	}
	reqConfigKeyMap := make(map[string]apistructs.PipelineCmsConfigKey, len(keys))
	for _, key := range keys {
		reqConfigKeyMap[key.Key] = key
	}
	cmsNs, exist, err := c.dbClient.GetCmsNs(pipelineSource, ns)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	configs, err := c.dbClient.GetCmsNsConfigs(cmsNs, transformKeysToStrSlice(keys...))
	if err != nil {
		return nil, err
	}
	result := make(map[string]apistructs.PipelineCmsConfigValue, len(configs))
	for _, config := range configs {
		// 默认使用 全局解密设置
		needDecrypt := globalDecrypt
		// 配置项级别的解密设置 覆盖 全局解密设置
		if reqConfig, ok := reqConfigKeyMap[config.Key]; ok {
			needDecrypt = reqConfig.Decrypt
		}
		// 在 db 中非加密存储不需要解密
		needDecrypt = needDecrypt && *config.Encrypt
		vv, err := c.decryptValueIfNeeded(needDecrypt, config.Value)
		if err != nil {
			return nil, err
		}

		// 加密存储且未解密的值是否展示
		if *config.Encrypt && !needDecrypt {
			// 默认不展示
			needShowEncryptedValue := false
			// 配置项级别的展示配置
			if reqConfig, ok := reqConfigKeyMap[config.Key]; ok {
				needShowEncryptedValue = reqConfig.ShowEncryptedValue
			}
			// 不需要展示，则 value 置空
			if !needShowEncryptedValue {
				vv = ""
			}
		}

		// 配置项级别的展示配置
		result[config.Key] = apistructs.PipelineCmsConfigValue{
			Value:       vv,
			EncryptInDB: *config.Encrypt,
			Type:        config.Type,
			Operations:  config.Extra.Operations,
			Comment:     config.Extra.Comment,
			From:        config.Extra.From,
			TimeCreated: config.TimeCreated,
			TimeUpdated: config.TimeUpdated,
		}
	}
	return result, nil
}

func getPipelineSourceFromContext(ctx context.Context) (apistructs.PipelineSource, error) {
	pipelineSource, ok := ctx.Value(CtxKeyPipelineSource).(apistructs.PipelineSource)
	if !ok || pipelineSource == "" {
		return "", errors.Errorf("missing %s", CtxKeyPipelineSource)
	}
	return pipelineSource, nil
}

func (c *pipelineCm) encryptValueIfNeeded(needEncrypt bool, origValue string) (string, error) {
	if !needEncrypt {
		return origValue, nil
	}
	encryptedV, err := c.rsaCrypt.Encrypt(origValue, encryption.Base64)
	if err != nil {
		return "", err
	}
	return encryptedV, nil
}

func (c *pipelineCm) decryptValueIfNeeded(needDecrypt bool, origValue string) (string, error) {
	if !needDecrypt {
		return origValue, nil
	}
	decryptedV, err := c.rsaCrypt.Decrypt(origValue, encryption.Base64)
	if err != nil {
		return "", err
	}
	return decryptedV, nil
}

func setDefault(v *apistructs.PipelineCmsConfigValue) {
	if v.Type == "" {
		v.Type = apistructs.PipelineCmsConfigTypeKV
	}
	if v.Operations == nil {
		v.Operations = &apistructs.PipelineCmsConfigDefaultOperationsForKV
		if v.Type == apistructs.PipelineCmsConfigTypeDiceFile {
			v.Operations.CanDownload = true
		}
	}
}

func validateConfigWhenUpdate(key string, v apistructs.PipelineCmsConfigValue) error {
	// key
	if err := strutil.Validate(key, strutil.NoChineseValidator, KeyValidator, EnvTransferValidator, strutil.MaxLenValidator(191)); err != nil {
		return err
	}
	// value
	if v.Type == "" {
		return errors.New("missing config type")
	}
	if !v.Type.Valid() {
		return errors.Errorf("invalid config type: %s", v.Type)
	}
	if err := strutil.Validate(key, strutil.EnvValueLenValidator); err != nil {
		return err
	}
	// comment
	if err := strutil.Validate(v.Comment, strutil.MaxLenValidator(200)); err != nil {
		return errors.Errorf("failed to validate comment, err: %v", err)
	}
	return nil
}

var KeyValidator strutil.Validator = func(s string) error {
	// 支持字母、数字、下划线、中划线、`.`，不能以数字、中划线、`.` 开头
	keyRegexp := `^[a-zA-Z_]+[.a-zA-Z0-9_-]*$`
	valid := regexp.MustCompilePOSIX(keyRegexp).MatchString(s)
	if !valid {
		return fmt.Errorf("valid key regexp: %s", keyRegexp)
	}
	return nil
}

var EnvTransferValidator strutil.Validator = func(s string) error {
	// 转换为 env
	env := strings.Replace(strings.Replace(strings.ToUpper(s), ".", "_", -1), "-", "_", -1)
	if err := strutil.Validate(env, strutil.EnvKeyValidator); err != nil {
		return errors.Errorf("failed to transfer config key to env key, config key: %s, env key: %s, err: %v",
			s, env, err)
	}
	return nil
}
