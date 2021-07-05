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

package mysql

import (
	"container/list"
	"fmt"
	"strings"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
	"github.com/erda-project/erda/pkg/discover"
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

	rootPassword := utils.GetRandomId()

	for _, service := range req.DiceYml.Services {
		// append envs
		envs := map[string]string{
			"ADDON_ID":            tmcInstance.ID,
			"ADDON_NODE_ID":       utils.GetRandomId(),
			"MYSQL_ROOT_PASSWORD": rootPassword,
		}
		utils.AppendMap(service.Envs, envs)

		// bind volumes
		hostPath := tmcInstance.ID
		serverId := service.Envs["SERVER_ID"]
		service.Binds = diceyml.Binds{
			clusterConfig["DICE_STORAGE_MOUNTPOINT"] + "/addon/mysql/backup/" + hostPath + "_" + serverId + ":/var/backup/mysql:rw",
			hostPath + "_" + serverId + ":/var/lib/mysql:rw",
		}

		// health check
		service.HealthCheck = diceyml.HealthCheck{
			Exec: &diceyml.ExecCheck{Cmd: fmt.Sprintf("mysql -uroot -p%s  -e 'select 1'", rootPassword)},
		}
	}

	req.GroupLabels["ADDON_GROUPS"] = "2" // what's this aim for

	return req
}

func (p *provider) DoPostDeployJob(tmcInstance *db.Instance, serviceGroupDeployResult interface{}, clusterConfig map[string]string) (map[string]string, error) {
	serviceGroup := serviceGroupDeployResult.(*apistructs.ServiceGroup)
	mysqlMap := parseResp2MySQLDtoMap(tmcInstance, serviceGroup)
	err := p.initMysql(mysqlMap, clusterConfig)
	if err != nil {
		return nil, err
	}

	time.Sleep(2 * time.Second)

	err = p.checkSalveStatus(mysqlMap, clusterConfig, err)
	if err != nil {
		return nil, err
	}

	resultConfig := map[string]string{
		"MYSQL_HOST":       mysqlMap["mysql"].mysqlHost,
		"MYSQL_PORT":       mysqlMap["mysql"].mysqlPort,
		"MYSQL_USERNAME":   "mysql",
		"MYSQL_PASSWORD":   mysqlMap["mysql"].password,
		"MYSQL_SLAVE_HOST": mysqlMap["mysql-slave"].mysqlHost,
		"MYSQL_SLAVE_PORT": mysqlMap["mysql-slave"].mysqlPort,
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

	// if custom resource, the database should created outside?
	if tmcInstance.IsCustom == "Y" {
		tenantConfig["MYSQL_DATABASES"] = tenantOptions["create_dbs"]
		return tenantConfig, nil
	}

	instanceConfig := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Config, &instanceConfig)

	mysqldto := mysqlDto{
		mysqlHost: instanceConfig["MYSQL_HOST"],
		mysqlPort: instanceConfig["MYSQL_PORT"],
		user:      instanceConfig["MYSQL_USERNAME"],
		password:  instanceConfig["MYSQL_PASSWORD"],
		options:   tenantOptions,
	}

	dbNames, err := p.createDb(mysqldto, clusterConfig, tenantConfig)
	if err != nil {
		return nil, err
	}

	err = p.initDb(dbNames, mysqldto, clusterConfig, err)
	if err != nil {
		return nil, nil
	}

	tenantConfig["MYSQL_DATABASES"] = strings.Join(dbNames, ",")
	return tenantConfig, nil
}

func (p *provider) initDb(dbNames []string, mysqldto mysqlDto, clusterConfig map[string]string, err error) error {
	if len(dbNames) == 0 {
		return nil
	}

	initSql := mysqldto.options["init_sql"]
	if len(initSql) == 0 {
		return nil
	}

	mysqlExec := &apistructs.MysqlExec{
		URL:      "jdbc:mysql://" + mysqldto.mysqlHost + ":" + mysqldto.mysqlPort,
		User:     mysqldto.user,
		Password: mysqldto.password,
	}

	mysqlExec.CreateDbs = dbNames
	mysqlExec.OssURL = initSql
	clusterInfo := apistructs.ClusterInfoData{}
	utils.JsonConvertObjToType(clusterConfig, &clusterInfo)
	err = p.Bdl.MySQLExecFile(mysqlExec, formatSoldierUrl(&clusterInfo))
	if err != nil {
		return err
	}
	return nil
}

type mysqlDto struct {
	mysqlHost string
	mysqlPort string
	user      string
	password  string
	options   map[string]string
	shortHost string
}

func parseResp2MySQLDtoMap(instance *db.Instance, serviceGroup *apistructs.ServiceGroup) map[string]*mysqlDto {
	resultMap := map[string]*mysqlDto{}

	options := map[string]string{}
	utils.JsonConvertObjToType(instance.Options, &options)

	for _, service := range serviceGroup.Services {
		resultMap[service.Name] = &mysqlDto{
			mysqlHost: service.Vip,
			mysqlPort: "3306",
			user:      "root",
			password:  options["MYSQL_ROOT_PASSWORD"],
			shortHost: service.ShortVIP,
			options:   options, // reuse same options
		}
	}

	return resultMap
}

func (p *provider) initMysql(mysqlMap map[string]*mysqlDto, clusterConfig map[string]string) error {
	password := mysqlMap["mysql"].password
	masterShortHost := mysqlMap["mysql"].shortHost

	linkList := list.New()
	for name, service := range mysqlMap {
		execDto := &apistructs.MysqlExec{
			URL:      "jdbc:mysql://" + service.mysqlHost + ":" + service.mysqlPort,
			User:     service.user,
			Password: service.password,
		}

		if name == "mysql" {
			execDto.Sqls = []string{strings.Replace(apistructs.AddonMysqlMasterGrantBackupSqls, "${MYSQL_ROOT_PASSWORD}", password, -1),
				strings.Replace(apistructs.AddonMysqlCreateMysqlUserSqls, "${MYSQL_ROOT_PASSWORD}", password, -1),
				apistructs.AddonMysqlGrantMysqlUserSqls,
				apistructs.AddonMysqlFlushSqls}
			linkList.PushFront(execDto) // master at first
		} else {
			execDto.Sqls = []string{strings.Replace(strings.Replace(apistructs.AddonMysqlSlaveChangeMasterSqls, "${MYSQL_ROOT_PASSWORD}", password, -1), "${MASTER_HOST}", masterShortHost, -1),
				apistructs.AddonMysqlSlaveResetSlaveSqls,
				apistructs.AddonMysqlSlaveStartSlaveSqls,
				strings.Replace(apistructs.AddonMysqlCreateMysqlUserSqls, "${MYSQL_ROOT_PASSWORD}", password, -1),
				apistructs.AddonMysqlGrantSelectMysqlUserSqls,
				apistructs.AddonMysqlFlushSqls}
			linkList.PushBack(execDto)
		}
	}
	var mysqlExecList []apistructs.MysqlExec
	for p := linkList.Front(); p != nil; p = p.Next() {
		mysqlExecList = append(mysqlExecList, *p.Value.(*apistructs.MysqlExec))
	}

	clusterInfo := apistructs.ClusterInfoData{}
	utils.JsonConvertObjToType(clusterConfig, &clusterInfo)
	err := p.Bdl.MySQLInit(&mysqlExecList, formatSoldierUrl(&clusterInfo))

	return err
}

func (p *provider) checkSalveStatus(mysqlMap map[string]*mysqlDto, clusterConfig map[string]string, err error) error {
	service := mysqlMap["mysql-slave"]
	mysqlExec := &apistructs.MysqlExec{
		URL:      "jdbc:mysql://" + service.mysqlHost + ":" + service.mysqlPort,
		User:     service.user,
		Password: service.password,
	}

	clusterInfo := apistructs.ClusterInfoData{}
	utils.JsonConvertObjToType(clusterConfig, &clusterInfo)
	err = p.Bdl.MySQLCheck(mysqlExec, formatSoldierUrl(&clusterInfo))
	if err != nil {
		return err
	}
	return nil
}

func (p *provider) createDb(mysqldto mysqlDto, clusterConfig map[string]string, tenantConfig map[string]string) ([]string, error) {
	mysqlExec := &apistructs.MysqlExec{
		URL:      "jdbc:mysql://" + mysqldto.mysqlHost + ":" + mysqldto.mysqlPort,
		User:     mysqldto.user,
		Password: mysqldto.password,
	}

	var createdDbNames []string
	createDbs := strings.Split(mysqldto.options["create_dbs"], ",")
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
	clusterInfo := apistructs.ClusterInfoData{}
	utils.JsonConvertObjToType(clusterConfig, &clusterInfo)
	err := p.Bdl.MySQLExec(mysqlExec, formatSoldierUrl(&clusterInfo))
	return createdDbNames, err
}

// formatSoldierUrl 拼接soldier地址
func formatSoldierUrl(clusterInfo *apistructs.ClusterInfoData) string {
	if (*clusterInfo)[apistructs.DICE_IS_EDGE] == "false" {
		return "http://" + discover.Soldier()
	}
	port := "80"
	protocol := "https"
	if strutil.Contains((*clusterInfo)[apistructs.DICE_PROTOCOL], "https") {
		port = (*clusterInfo)[apistructs.DICE_HTTPS_PORT]
	} else {
		protocol = "http"
		port = (*clusterInfo)[apistructs.DICE_HTTP_PORT]
	}
	return protocol + "://soldier." + (*clusterInfo)[apistructs.DICE_ROOT_DOMAIN] + ":" + port

}
