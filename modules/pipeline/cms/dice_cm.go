package cms

import (
	"context"
	"errors"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	defaultProjectID = 10000000

	errNSNotExist = "namespace not exist"
)

type diceCM struct {
	bdl      *bundle.Bundle
	nsPrefix string
}

func NewDiceCM(nsPrefix string) *diceCM {
	bdl := bundle.New(bundle.WithCMDB())
	var diceCM diceCM
	diceCM.bdl = bdl
	diceCM.nsPrefix = nsPrefix
	return &diceCM
}

func (d *diceCM) IdempotentCreateNS(ctx context.Context, ns string) error {
	var req apistructs.NamespaceCreateRequest
	req.Name = d.wrapNS(ns)
	req.Dynamic = false
	req.ProjectID = defaultProjectID
	return d.bdl.CreateNamespace(req)
}

func (d *diceCM) IdempotentDeleteNS(ctx context.Context, ns string) error {
	err := d.bdl.DeleteNamespace(d.wrapNS(ns))
	if err != nil && strings.Contains(err.Error(), errNSNotExist) {
		return nil
	}
	return err
}

func (d *diceCM) PrefixListNS(ctx context.Context, nsPrefix string) ([]apistructs.PipelineCmsNs, error) {
	panic("not implemented")
}

func (d *diceCM) UpdateConfigs(ctx context.Context, ns string, kvs map[string]apistructs.PipelineCmsConfigValue) error {
	if err := d.IdempotentCreateNS(ctx, ns); err != nil {
		return err
	}
	configItems := make([]apistructs.EnvConfig, 0, len(kvs))
	for k, v := range kvs {
		configItems = append(configItems, apistructs.EnvConfig{Key: k, Value: v.Value, Encrypt: v.EncryptInDB})
	}
	return d.bdl.AddOrUpdateNamespaceConfig(d.wrapNS(ns), configItems, true)
}

func (d *diceCM) DeleteConfigs(ctx context.Context, ns string, keys ...string) error {
	existKVs, err := d.GetConfigs(ctx, ns, false, transformStrSliceToKeys(true, keys...)...)
	if err != nil {
		return err
	}
	var keysNeedDelete []string
	for _, k := range keys {
		if _, ok := existKVs[k]; ok {
			keysNeedDelete = append(keysNeedDelete, k)
		}
	}
	var errs []string
	for _, delKey := range keysNeedDelete {
		if err := d.bdl.DeleteNamespaceConfig(d.wrapNS(ns), delKey); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strutil.Join(errs, ", ", true))
}

// GetConfigs 若 ns 不存在，则会自动创建
func (d *diceCM) GetConfigs(ctx context.Context, ns string, globalDecrypt bool, keys ...apistructs.PipelineCmsConfigKey) (map[string]apistructs.PipelineCmsConfigValue, error) {
	kvs, err := d.bdl.FetchNamespaceConfig(apistructs.EnvConfigFetchRequest{
		Namespace:            d.wrapNS(ns),
		Decrypt:              globalDecrypt,
		AutoCreateIfNotExist: true,
		CreateReq: apistructs.NamespaceCreateRequest{
			Name:      d.wrapNS(ns),
			ProjectID: defaultProjectID,
			Dynamic:   false,
		},
	})
	if err != nil {
		return nil, err
	}
	result := make(map[string]apistructs.PipelineCmsConfigValue, len(kvs))
	for k, v := range kvs {
		result[k] = apistructs.PipelineCmsConfigValue{Value: v, EncryptInDB: true}
	}
	// 过滤
	if len(keys) > 0 {
		filterResult := make(map[string]apistructs.PipelineCmsConfigValue)
		for _, filterKey := range keys {
			if v, ok := result[filterKey.Key]; ok {
				filterResult[filterKey.Key] = v
			}
		}
		return filterResult, nil
	}
	return result, nil
}

// wrapNS 在请求的 ns 基础上加上前缀。
// dice cm 创建 ns 时需要 projectID，为了兼容没有 projectID 的情况，强制 projectID = 1000w。
// 为了避免和 projectID 真为 1000w 的数据冲突，通过本库创建的 ns 会自动加上前缀用于区分。
func (d *diceCM) wrapNS(ns string) string {
	return d.nsPrefix + ns
}
