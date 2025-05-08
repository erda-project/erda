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

package mysql

import (
	"container/list"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	conf "github.com/erda-project/erda/cmd/erda-server/conf/msp"
	"github.com/erda-project/erda/internal/apps/msp/instance/db"
	"github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/internal/apps/msp/resource/utils"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/addon"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/mysqlhelper"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourceMysql
}

func (p *provider) CheckIfHasCustomConfig(clusterConfig map[string]string) (map[string]string, bool) {
	// try find if aliyun mse instance exists, reuse it if any
	mysqlHost, ok := clusterConfig["MS_MYSQL_HOST"]
	if !ok {
		return nil, false
	}

	mysqlPort, ok := clusterConfig["MS_MYSQL_PORT"]
	if !ok {
		return nil, false
	}

	mysqlUser, ok := clusterConfig["MS_MYSQL_USER"]
	if !ok {
		return nil, false
	}

	mysqlPassword, ok := clusterConfig["MS_MYSQL_PASSWORD"]
	if !ok {
		return nil, false
	}
	decryptPassword := utils.AesDecrypt(mysqlPassword)

	config := map[string]string{
		"MYSQL_HOST":       mysqlHost,
		"MYSQL_PORT":       mysqlPort,
		"MYSQL_USERNAME":   mysqlUser,
		"MYSQL_PASSWORD":   decryptPassword,
		"MYSQL_SLAVE_HOST": mysqlHost,
		"MYSQL_SLAVE_PORT": mysqlPort,
	}

	return config, true
}

func (p *provider) BuildServiceGroupRequest(resourceInfo *handlers.ResourceInfo, tmcInstance *db.Instance, clusterConfig map[string]string) interface{} {
	req := p.DefaultDeployHandler.BuildServiceGroupRequest(resourceInfo, tmcInstance, clusterConfig).(*apistructs.ServiceGroupCreateV2Request)

	rootPassword := uuid.UUID()

	for _, service := range req.DiceYml.Services {
		// append envs
		envs := map[string]string{
			"ADDON_ID":            tmcInstance.ID,
			"ADDON_NODE_ID":       utils.GetRandomId(),
			"MYSQL_ROOT_PASSWORD": rootPassword,
		}
		utils.AppendMap(service.Envs, envs)

		//labels
		if service.Labels == nil {
			service.Labels = make(map[string]string)
		}
		options := map[string]string{}
		utils.JsonConvertObjToType(tmcInstance.Options, &options)
		utils.SetlabelsFromOptions(options, service.Labels)

		// bind volumes
		/*
			hostPath := tmcInstance.ID
			serverId := service.Envs["SERVER_ID"]
			service.Binds = diceyml.Binds{
				clusterConfig["DICE_STORAGE_MOUNTPOINT"] + "/addon/mysql/backup/" + hostPath + "_" + serverId + ":/var/backup/mysql:rw",
				hostPath + "_" + serverId + ":/var/lib/mysql:rw",
			}
		*/
		//  /var/backup/mysql volume
		vol01 := addon.SetAddonVolumes(options, "/var/backup/mysql", false)
		//  /var/lib/mysql volume
		vol02 := addon.SetAddonVolumes(options, "/var/lib/mysql", false)
		service.Volumes = diceyml.Volumes{vol01, vol02}

		// health check
		service.HealthCheck = diceyml.HealthCheck{
			Exec: &diceyml.ExecCheck{Cmd: fmt.Sprintf("mysql -uroot -p%s  -e 'select 1'", rootPassword)},
		}
	}

	req.GroupLabels["ADDON_GROUPS"] = "2" // what's this aim for

	return req
}

func (h *provider) DoDeploy(serviceGroupDeployRequest interface{}, resourceInfo *handlers.ResourceInfo, tmcInstance *db.Instance, clusterConfig map[string]string) (
	interface{}, error) {

	// persistent the root password
	serviceGroup := serviceGroupDeployRequest.(*apistructs.ServiceGroupCreateV2Request)
	masterService := serviceGroup.DiceYml.Services["mysql"]
	password := masterService.Envs["MYSQL_ROOT_PASSWORD"]

	options := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Options, options)
	options["MYSQL_ROOT_PASSWORD"] = password
	optionsStr, _ := utils.JsonConvertObjToString(options)
	if err := h.InstanceDb.Model(tmcInstance).Update("options", optionsStr).Error; err != nil {
		return nil, err
	}

	return h.DefaultDeployHandler.DoDeploy(serviceGroupDeployRequest, resourceInfo, tmcInstance, clusterConfig)
}

func (h *provider) CheckIfNeedTmcInstance(req *handlers.ResourceDeployRequest, resourceInfo *handlers.ResourceInfo) (*db.Instance, bool, error) {
	// mysql remove the `version` condition. because in the old cluster nacos[1.1.0] depend on mysql[5.7] but now depend on mysql[8.0]
	var where = map[string]any{
		"engine":     resourceInfo.TmcVersion.Engine,
		"az":         req.Az,
		"status":     handlers.TmcInstanceStatusRunning,
		"is_deleted": apistructs.AddonNotDeleted,
	}
	if resourceInfo.TmcVersion.Version != "" {
		where["version"] = resourceInfo.TmcVersion.Version
		logrus.Infof("[mysql] check if need tmc instance, version: %s", where["version"])
	}
	instance, ok, err := h.InstanceDb.First(where)
	if err != nil {
		return nil, false, err
	}
	return instance, !ok, nil
}

func (p *provider) DoPostDeployJob(tmcInstance *db.Instance, serviceGroupDeployResult interface{}, clusterConfig map[string]string) (map[string]string, error) {
	serviceGroup := serviceGroupDeployResult.(*apistructs.ServiceGroup)
	mysqlMap := ParseResp2MySQLDtoMap(tmcInstance, serviceGroup)
	var deployByOperator = serviceGroup.Labels["USE_OPERATOR"] == "mysql"
	if err := p.initMysql(mysqlMap, clusterConfig, deployByOperator); err != nil {
		p.Log.Infof("failed to initMySQL, mysqlMqp: %s, clusterConfig: %s, tmc_instance: %s, serviceGroup: %s, err: %v",
			strutil.TryGetJsonStr(mysqlMap), strutil.TryGetJsonStr(clusterConfig), strutil.TryGetJsonStr(tmcInstance), strutil.TryGetJsonStr(serviceGroup), err)
		return nil, errors.Wrap(err, "failed to initMySQL")
	}

	time.Sleep(2 * time.Second)

	if !deployByOperator {
		if err := p.checkSalveStatus(mysqlMap, clusterConfig); err != nil {
			p.Log.Errorf("failed to checkSalveStatus, mysqlMap: %s, clusterConfig: %s, tmc_instance: %s, serviceGroup: %s, err: %v",
				strutil.TryGetJsonStr(mysqlMap), strutil.TryGetJsonStr(clusterConfig), strutil.TryGetJsonStr(tmcInstance), strutil.TryGetJsonStr(serviceGroup), err)
			return nil, errors.Wrap(err, "failed to checkSalveStatus")
		}
	}

	var resultConfig = make(map[string]string)
	if dto := mysqlMap["mysql"]; dto != nil {
		resultConfig["MYSQL_HOST"] = dto.MySQLHost
		resultConfig["MYSQL_PORT"] = dto.MySQLPort
		resultConfig["MYSQL_USERNAME"] = "mysql"
		resultConfig["MYSQL_PASSWORD"] = dto.Password
	}
	if dto := mysqlMap["mysql-slave"]; dto != nil {
		resultConfig["MYSQL_SLAVE_HOST"] = dto.MySQLHost
		resultConfig["MYSQL_SLAVE_PORT"] = dto.MySQLPort
	}

	return resultConfig, nil
}

func (p *provider) BuildTmcInstanceConfig(tmcInstance *db.Instance, serviceGroupDeployResult interface{},
	clusterConfig map[string]string, additionalConfig map[string]string) map[string]string {
	return additionalConfig
}

func (p *provider) DoApplyTmcInstanceTenant(req *handlers.ResourceDeployRequest, resourceInfo *handlers.ResourceInfo,
	tmcInstance *db.Instance, tenant *db.InstanceTenant, clusterConfig map[string]string) (map[string]string, error) {

	tenantConfig := map[string]string{}

	tenantOptions := map[string]string{}
	utils.JsonConvertObjToType(tenant.Options, &tenantOptions)
	tenantConfig["MYSQL_DATABASES"] = tenantOptions["create_dbs"]

	// if custom resource, the database should created outside?
	if tmcInstance.IsCustom == "Y" {
		return tenantConfig, nil
	}

	instanceConfig := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Config, &instanceConfig)

	mysqldto := mysqlDto{
		MySQLHost: instanceConfig["MYSQL_HOST"],
		MySQLPort: instanceConfig["MYSQL_PORT"],
		User:      instanceConfig["MYSQL_USERNAME"],
		Password:  instanceConfig["MYSQL_PASSWORD"],
		Options:   tenantOptions,
	}

	dbNames, err := p.createDb(mysqldto, clusterConfig, tenantConfig)
	if err != nil {
		return nil, err
	}

	err = p.initDb(dbNames, mysqldto, clusterConfig, err)
	if err != nil {
		return nil, err
	}

	tenantConfig["MYSQL_DATABASES"] = strings.Join(dbNames, ",")
	return tenantConfig, nil
}

func (p *provider) initDb(dbNames []string, mysqldto mysqlDto, clusterConfig map[string]string, err error) error {
	if len(dbNames) == 0 {
		return nil
	}

	initSql := mysqldto.Options["init_sql"]
	if len(initSql) == 0 {
		return nil
	}

	mysqlExec := &mysqlhelper.Request{
		ClusterKey: clusterConfig["DICE_CLUSTER_NAME"],
		Url:        "jdbc:mysql://" + mysqldto.MySQLHost + ":" + mysqldto.MySQLPort,
		User:       mysqldto.User,
		Password:   mysqldto.Password,
		CreateDbs:  dbNames,
	}

	// Check if the script has been executed
	p.Log.Infof("[%s] to check if the SQL script has been executed", initSql)
	ok, err := mysqlExec.HasRecord(initSql)
	if err != nil {
		return errors.Wrapf(err, "failed to check the init sql record for %s", initSql)
	}
	if ok {
		p.Log.Infof("[%s] there is already a record for the SQL script, skip this initialization", initSql)
		return nil
	}

	p.Log.Infof("[%s] the SQL script has not been executed, to execute it", initSql)
	sql, err := p.tryReadFile(initSql)
	if err != nil {
		return err
	}
	mysqlExec.Sqls = []string{sql}

	if err = mysqlExec.Exec(); err != nil {
		return err
	}

	// Logging script execution record
	p.Log.Infof("[%s] to record the SQL script execution history", initSql)
	return mysqlExec.Record(initSql)
}

type mysqlDto struct {
	MySQLHost string            `json:"mySQLHost"`
	MySQLPort string            `json:"mySQLPort"`
	User      string            `json:"user"`
	Password  string            `json:"password"`
	Options   map[string]string `json:"options"`
	ShortHost string            `json:"shortHost"`
}

func ParseResp2MySQLDtoMap(instance *db.Instance, serviceGroup *apistructs.ServiceGroup) map[string]*mysqlDto {
	resultMap := map[string]*mysqlDto{}

	options := map[string]string{}
	utils.JsonConvertObjToType(instance.Options, &options)

	for _, service := range serviceGroup.Services {
		resultMap[service.Name] = &mysqlDto{
			MySQLHost: service.Vip,
			MySQLPort: "3306",
			User:      "root",
			Password:  options["MYSQL_ROOT_PASSWORD"],
			ShortHost: service.ShortVIP,
			Options:   options, // reuse same options
		}
	}

	return resultMap
}

// create mysql account "mysql" and grant users
func (p *provider) initMysql(mysqlMap map[string]*mysqlDto, clusterConfig map[string]string, deployByOperator bool) error {
	password := mysqlMap["mysql"].Password
	masterShortHost := mysqlMap["mysql"].ShortHost
	p.Log.Infof("mysqlMap: %s", strutil.TryGetJsonStr(mysqlMap))

	linkList := list.New()
	for name, service := range mysqlMap {
		var configs = []Option{
			WithUseOperator(deployByOperator),
			WithClusterKey(clusterConfig["DICE_CLUSTER_NAME"]),
			WithAddress(service.MySQLHost + ":" + service.MySQLPort),
			WithUsername(service.User),
			WithPassword(service.Password),
			WithOperatorCli(p.MyOperaInsCli),
		}
		if name == "mysql" {
			var item = NewInstanceAdapter(append(configs, WithQueries([]string{
				strings.Replace(apistructs.AddonMysqlMasterGrantBackupSqls, "${MYSQL_ROOT_PASSWORD}", password, -1),
				strings.Replace(apistructs.AddonMysqlCreateMysqlUserSqls, "${MYSQL_ROOT_PASSWORD}", password, -1),
				apistructs.AddonMysqlGrantMysqlUserSqls,
				apistructs.AddonMysqlFlushSqls,
			}))...)
			linkList.PushFront(item)
		} else {
			var item = NewInstanceAdapter(append(configs, WithQueries([]string{
				strings.Replace(strings.Replace(apistructs.AddonMysqlSlaveChangeMasterSqls, "${MYSQL_ROOT_PASSWORD}", password, -1), "${MASTER_HOST}", masterShortHost, -1),
				apistructs.AddonMysqlSlaveResetSlaveSqls,
				apistructs.AddonMysqlSlaveStartSlaveSqls,
				strings.Replace(apistructs.AddonMysqlCreateMysqlUserSqls, "${MYSQL_ROOT_PASSWORD}", password, -1),
				apistructs.AddonMysqlGrantSelectMysqlUserSqls,
				apistructs.AddonMysqlFlushSqls,
			}))...)
			linkList.PushBack(item)
		}
	}

	for e := linkList.Front(); e != nil; e = e.Next() {
		if err := e.Value.(InstanceAdapter).ExecSQLs(); err != nil {
			p.Log.Errorf("failed to Exec SQLs for %s", strutil.TryGetJsonStr(e.Value))
			return errors.Wrap(err, "failed to Exec SQLs")
		}
	}

	return nil
}

func (p *provider) checkSalveStatus(mysqlMap map[string]*mysqlDto, clusterConfig map[string]string) error {
	service := mysqlMap["mysql-slave"]
	mysqlExec := &mysqlhelper.Request{
		ClusterKey: clusterConfig["DICE_CLUSTER_NAME"],
		Url:        "jdbc:mysql://" + service.MySQLHost + ":" + service.MySQLPort,
		User:       service.User,
		Password:   service.Password,
	}

	status, err := mysqlExec.GetSlaveState()
	if err != nil {
		p.Log.Errorf("failed to GetSlaveState, MySQL request: %s, err: %v", strutil.TryGetJsonStr(mysqlExec), err)
		return errors.Wrap(err, "failed to GetSlaveState, MySQL request")
	}

	if !strings.EqualFold(status.IORunning, "connecting") &&
		!strings.EqualFold(status.IORunning, "yes") ||
		!strings.EqualFold(status.SQLRunning, "connecting") &&
			!strings.EqualFold(status.SQLRunning, "yes") {
		return fmt.Errorf("slave in error status")
	}
	return nil
}

func (p *provider) createDb(mysqldto mysqlDto, clusterConfig map[string]string, tenantConfig map[string]string) ([]string, error) {
	mysqlExec := &mysqlhelper.Request{
		ClusterKey: clusterConfig["DICE_CLUSTER_NAME"],
		Url:        "jdbc:mysql://" + mysqldto.MySQLHost + ":" + mysqldto.MySQLPort,
		User:       mysqldto.User,
		Password:   mysqldto.Password,
	}

	var createdDbNames []string
	createDbs := strings.Split(mysqldto.Options["create_dbs"], ",")
	var sqls []string
	for _, createDb := range createDbs {
		dbName := strings.Trim(createDb, " ")
		if len(dbName) == 0 {
			continue
		}
		sqls = append(sqls, "create database if not exists `"+dbName+"`;")
		createdDbNames = append(createdDbNames, dbName)
	}
	if len(sqls) == 0 {
		return createdDbNames, nil
	}
	mysqlExec.Sqls = sqls

	err := mysqlExec.Exec()
	return createdDbNames, err
}

func (p *provider) tryReadFile(file string) (string, error) {

	if !strings.HasPrefix(file, "file://") {
		return "", fmt.Errorf("not supported file storage type")
	}

	formattedPath := strings.TrimPrefix(file, "file://")
	formattedPath = strings.ReplaceAll(formattedPath, ".tar.gz", ".sql")
	formattedPath = conf.MSPAddonFsRootPath + "/" + formattedPath

	fs := conf.MSPAddonInitSqls

	data, err := fs.ReadFile(formattedPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
