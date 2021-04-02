package environment

import (
	"regexp"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/services/permission"
)

// EnvConfig 命名空间参数
type EnvConfig struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
}

// Option 定义 EnvConfig 对象的配置选项
type Option func(*EnvConfig)

// New 新建 EnvConfig 实例
func New(options ...Option) *EnvConfig {
	o := &EnvConfig{}
	for _, op := range options {
		op(o)
	}
	return o
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(o *EnvConfig) {
		o.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(c *EnvConfig) {
		c.bdl = bdl
	}
}

const (
	Web             = "WEB"
	DeployEnvFormat = "^[A-Za-z_][A-Za-z0-9_]*"
	NotDeleteValue  = "N"
)

// Add 添加 env config
func (e *EnvConfig) Add(createReq *apistructs.EnvConfigAddOrUpdateRequest, namespace string, encrypt bool) error {
	// check params
	err := verifyEnvConfigs(createReq.Configs)
	if err != nil {
		return err
	}

	// check namespace if exist
	ns, err := e.db.GetNamespaceByName(namespace)
	if err != nil {
		return err
	}

	if ns == nil {
		return errors.Errorf("not exist namespace, namespace: %s", namespace)
	}

	configItems := encryptAndParse2Entity(createReq.Configs, ns, encrypt)

	for _, config := range configItems {
		err = e.db.UpdateOrAddEnvConfig(&config)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update 更新 env config
func (e *EnvConfig) Update(permission *permission.Permission, createReq *apistructs.EnvConfigAddOrUpdateRequest, namespace, userID string, encrypt bool) error {
	// check params
	err := verifyEnvConfigs(createReq.Configs)
	if err != nil {
		return err
	}

	// check namespace if exist
	ns, err := e.db.GetNamespaceByName(namespace)
	if err != nil {
		return err
	}

	if ns == nil {
		return errors.Errorf("not exist namespace, namespace: %s", namespace)
	}

	// TODO 操作鉴权
	//appID, err := strconv.ParseUint(ns.ApplicationID, 10, 64)
	//if err != nil {
	//	return err
	//}
	//req := apistructs.PermissionCheckRequest{
	//	UserID:   userID,
	//	Scope:    apistructs.AppScope,
	//	ScopeID:  appID,
	//	Resource: apistructs.AppResource,
	//	Action:   apistructs.DeleteAction,
	//}
	//
	//if access, err := permission.CheckPermission(&req); err != nil || !access {
	//	return apierrors.ErrUpdateEnvConfig.AccessDenied()
	//}

	configItems := encryptAndParse2Entity(createReq.Configs, ns, encrypt)

	for _, config := range configItems {
		configItem, err := e.db.GetEnvConfigByKey(ns.ID, config.ItemKey)
		if err != nil {
			return errors.Errorf("failed to get config by namespace id and key, namespaceID: %s, key: %s",
				namespace, config.ItemKey)
		}

		if configItem != nil {
			config.ID = configItem.ID
			config.CreatedAt = configItem.CreatedAt
		}

		err = e.db.UpdateOrAddEnvConfig(&config)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetConfigs 根据指定 namespace 获取 env config
func (e *EnvConfig) GetConfigs(permission *permission.Permission, namespace, userID string, decrypt bool) ([]apistructs.EnvConfig, error) {
	// check namespace if exist
	ns, err := e.db.GetNamespaceByName(namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get namespace by name")
	}

	if ns == nil {
		return nil, errors.Errorf("not exist namespace, namespace: %s", namespace)
	}

	// TODO 操作鉴权
	//appID, err := strconv.ParseUint(ns.ApplicationID, 10, 64)
	//if err != nil {
	//	return nil, err
	//}
	//req := apistructs.PermissionCheckRequest{
	//	UserID:   userID,
	//	Scope:    apistructs.AppScope,
	//	ScopeID:  appID,
	//	Resource: apistructs.AppResource,
	//	Action:   apistructs.GetAction,
	//}
	//
	//if access, err := permission.CheckPermission(&req); err != nil || !access {
	//	return nil, apierrors.ErrGetNamespaceEnvConfig.AccessDenied()
	//}

	configItems, err := e.db.GetEnvConfigsByNamespaceID(ns.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get configs by namespace id")
	}

	if configItems == nil {
		return nil, nil
	}

	return decryptAndEntitys2Res(configItems, decrypt), nil
}

// DeleteConfig 删除指定 namespace 下的某个配置
func (e *EnvConfig) DeleteConfig(permission *permission.Permission, namespace, key, userID string) error {
	// check namespace if exist
	ns, err := e.db.GetNamespaceByName(namespace)
	if err != nil {
		return err
	}

	if ns == nil {
		return errors.Errorf("not exist namespace, namespace: %s", namespace)
	}

	// TODO 操作鉴权
	//appID, err := strconv.ParseUint(ns.ApplicationID, 10, 64)
	//if err != nil {
	//	return err
	//}
	//req := apistructs.PermissionCheckRequest{
	//	UserID:   userID,
	//	Scope:    apistructs.AppScope,
	//	ScopeID:  appID,
	//	Resource: apistructs.AppResource,
	//	Action:   apistructs.DeleteAction,
	//}
	//
	//if access, err := permission.CheckPermission(&req); err != nil || !access {
	//	return apierrors.ErrDeleteEnvConfig.AccessDenied()
	//}

	configItem, err := e.db.GetEnvConfigByKey(ns.ID, key)
	if err != nil {
		return err
	}

	if configItem == nil {
		return errors.New("not exist config")
	}

	err = e.db.SoftDeleteEnvConfig(configItem)
	return err
}

// GetMultiNamespaceConfigs 根据多个 namespace 获取所有配置信息
func (e *EnvConfig) GetMultiNamespaceConfigs(permission *permission.Permission, userID string, namespaceParams []apistructs.NamespaceParam) (map[string][]apistructs.EnvConfig, error) {
	// check namespace params
	if namespaceParams == nil {
		return nil, errors.New("namespace param is nil")
	}

	if len(namespaceParams) > 10 {
		return nil, errors.New("namespace num too long")
	}

	mapEnvConfigs := make(map[string][]apistructs.EnvConfig)
	for _, nsp := range namespaceParams {
		config, err := e.GetConfigs(permission, nsp.NamespaceName, userID, nsp.Decrypt)
		if err != nil {
			return nil, err
		}
		mapEnvConfigs[nsp.NamespaceName] = config
	}

	for k := range mapEnvConfigs {
		for i := range mapEnvConfigs[k] {
			mapEnvConfigs[k][i].Type = map[string]string{"ENV": "kv", "FILE": "dice-file"}[mapEnvConfigs[k][i].ConfigType]
			if mapEnvConfigs[k][i].Type == "dice-file" {
				mapEnvConfigs[k][i].Operations = &apistructs.PipelineCmsConfigOperations{
					CanDownload: true,
					CanEdit:     true,
					CanDelete:   true,
				}
			} else if mapEnvConfigs[k][i].Operations == nil {
				mapEnvConfigs[k][i].Operations = &apistructs.PipelineCmsConfigOperations{
					CanDownload: false,
					CanEdit:     true,
					CanDelete:   true,
				}
			}
		}
	}

	return mapEnvConfigs, nil
}

// GetDeployConfigs 根据指定 namespace 获取部署环境变量配置
func (e *EnvConfig) GetDeployConfigs(permission *permission.Permission, userID, namespace string) ([]apistructs.EnvConfig, error) {
	// check namespace if exist
	ns, err := e.db.GetNamespaceByName(namespace)
	if err != nil {
		return nil, err
	}

	if ns == nil {
		return nil, errors.Errorf("not exist namespace, namespace: %s", namespace)
	}

	// TODO 操作鉴权
	//appID, err := strconv.ParseUint(ns.ApplicationID, 10, 64)
	//if err != nil {
	//	return nil, err
	//}
	//req := apistructs.PermissionCheckRequest{
	//	UserID:   userID,
	//	Scope:    apistructs.AppScope,
	//	ScopeID:  appID,
	//	Resource: apistructs.AppResource,
	//	Action:   apistructs.GetAction,
	//}
	//
	//if access, err := permission.CheckPermission(&req); err != nil || !access {
	//	return nil, apierrors.ErrGetDeployEnvConfig.AccessDenied()
	//}

	configItems, err := e.db.GetEnvConfigsByNamespaceID(ns.ID)
	if err != nil {
		return nil, err
	}

	// merge default namespace configs
	if !ns.IsDefault {
		nsRelation, err := e.db.GetNamespaceRelationByName(ns.Name)
		if err != nil {
			return nil, err
		}

		if nsRelation != nil {
			defaultNs, err := e.db.GetNamespaceByName(nsRelation.DefaultNamespace)
			if err != nil {
				return nil, err
			}

			envCongigs, err := e.db.GetEnvConfigsByNamespaceID(defaultNs.ID)
			if err != nil {
				return nil, err
			}

			configItems = append(configItems, envCongigs...)
		}
	}

	newConfigsItem, err := filterDeployEnvFormat(configItems)
	if err != nil {
		return nil, err
	}

	return decryptAndEntitys2Res(newConfigsItem, true), nil
}

func filterDeployEnvFormat(configs []model.ConfigItem) ([]model.ConfigItem, error) {
	var newConfigsItem []model.ConfigItem
	configMap := make(map[string]struct{}, 0)
	for _, env := range configs {
		m, err := regexp.MatchString(DeployEnvFormat, env.ItemKey)
		if err != nil {
			return nil, errors.Errorf("failed to match config key, namaspace: %s, parten: %s, (%+v)",
				env.ItemKey, DeployEnvFormat, err)
		}
		if m {
			if _, ok := configMap[env.ItemKey]; !ok {
				configMap[env.ItemKey] = struct{}{}
				newConfigsItem = append(newConfigsItem, env)
			}
		}
	}

	return newConfigsItem, nil
}

func verifyEnvConfigs(configs []apistructs.EnvConfig) error {
	if configs == nil {
		return errors.New("body param env config is null")
	}

	for i, env := range configs {
		if env.Key == "" {
			return errors.Errorf("environment config key error, env: %+v", env)
		}
		if env.Value == "" {
			return errors.Errorf("environment config key error, env: %+v", env)
		}
		if env.ConfigType != "FILE" && env.ConfigType != "ENV" {
			configs[i].ConfigType = "ENV"
		}
	}
	return nil
}

func encryptAndParse2Entity(configs []apistructs.EnvConfig, namespace *model.ConfigNamespace, encrypt bool) []model.ConfigItem {
	if configs == nil {
		return nil
	}

	configItems := []model.ConfigItem{}
	for _, config := range configs {
		configItem := model.ConfigItem{}
		configItem.ItemKey = config.Key
		configItem.ItemValue = config.Value
		configItem.ItemType = config.ConfigType
		if encrypt {
			configItem.Encrypt = true
		}
		configItem.Source = Web
		configItem.ItemComment = config.Comment
		configItem.NamespaceID = uint64(namespace.ID)
		configItem.Dynamic = namespace.Dynamic
		configItem.IsDeleted = NotDeleteValue

		configItems = append(configItems, configItem)
	}

	return configItems
}

func decryptAndEntitys2Res(configItems []model.ConfigItem, decrypt bool) []apistructs.EnvConfig {
	envConfigs := []apistructs.EnvConfig{}
	for _, config := range configItems {
		envConfig := apistructs.EnvConfig{
			Key:        config.ItemKey,
			Comment:    config.ItemComment,
			Status:     config.Status,
			Encrypt:    config.Encrypt,
			Source:     config.Source,
			CreateTime: config.CreatedAt,
			UpdateTime: config.UpdatedAt,
			ConfigType: config.ItemType,
		}

		if decrypt || !config.Encrypt {
			envConfig.Value = config.ItemValue
		}

		envConfigs = append(envConfigs, envConfig)
	}

	return envConfigs
}
