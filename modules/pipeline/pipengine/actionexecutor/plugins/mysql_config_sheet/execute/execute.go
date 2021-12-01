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

package execute

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/actionagent"
	"github.com/erda-project/erda/modules/pipeline/pkg/jsonparse"
	"github.com/erda-project/erda/modules/pipeline/pkg/log_collector"
	"github.com/erda-project/erda/modules/pipeline/pkg/pipelinefunc"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/apitestsv2"
	jsonfilter "github.com/erda-project/erda/pkg/encoding/jsonparse"
	"github.com/erda-project/erda/pkg/envconf"
	"github.com/erda-project/erda/pkg/loop"
)

const (
	mysqlHost     = "MYSQL_HOST"
	mysqlPassword = "MYSQL_PASSWORD"
	mysqlPort     = "MYSQL_PORT"
	mysqlUsername = "MYSQL_USERNAME"

	MetaAssertResultKey = "assert_result"

	ResultSuccess = "success"
	ResultFailed  = "failed"
)

func Do(ctx context.Context, task *spec.PipelineTask, bdl *bundle.Bundle) {
	logger := log_collector.NewLogger().WithContext(context.WithValue(context.Background(), log_collector.CtxKeyCollectorLogID, task.Extra.UUID))
	ctx = context.WithValue(ctx, log_collector.CtxKeyLogger, logger)

	meta := NewMeta()
	meta.AssertResult = true
	var err error

	defer writeMetaFile(ctx, task, meta)
	defer func() {
		if err != nil {
			log_collector.Clog(ctx).Errorf(err.Error())
			logrus.Errorf("mysql_config_sheet exec error %v", err)
			meta.Err = err
			meta.AssertResult = false
		}
	}()

	// parse conf from task
	cfg, err := parseConfFromTask(task)
	if err != nil {
		err = fmt.Errorf("failed to parse config from task, err: %v", err)
		return
	}

	req, err := genRequest(cfg, bdl)
	if err != nil {
		err = fmt.Errorf("failed to genRequest, err: %v", err)
		return
	}
	req.ClusterKey = task.Extra.ClusterName

	log_collector.Clog(ctx).Infof("----------- execute sql ----------- ")

	db, err := req.dbOpen()
	if err != nil {
		err = fmt.Errorf("dbOpen error %v", err)
		return
	}
	defer func() {
		_ = db.Close()
	}()

	if cfg.Database != "" {
		cfg.Executors = append([]SqlExecutor{
			{
				Sql: "USE " + cfg.Database,
			},
		}, cfg.Executors...)
	}

	err = exec(ctx, db, meta, cfg.Executors)
	if err != nil {
		err = fmt.Errorf("sql exec error %v", err)
		return
	}

	return
}

func exec(ctx context.Context, db *sql.DB, meta *Meta, executors []SqlExecutor) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("db translate begin error %v", err)
	}

	var allResults = []interface{}{}
	defer func() {
		meta.Result = jsonfilter.JsonOneLine(allResults)
	}()
	meta.OutParamsResult = map[string]interface{}{}
	meta.OutParams = []apistructs.APIOutParam{}
	for _, exec := range executors {
		log_collector.Clog(ctx).Infof("sql：%v", exec.Sql)
		result, err := runSql(exec.Sql, tx)
		if err != nil {
			return err
		}
		log_collector.Clog(ctx).Infof("result： %v", jsonfilter.JsonOneLine(result))
		allResults = append(allResults, result)

		outParamsResult := map[string]interface{}{}
		outParams := []apistructs.APIOutParam{}
		for _, outputParam := range exec.OutParams {
			output := jsonfilter.FilterJson([]byte(jsonfilter.JsonOneLine(result)), outputParam.Expression, "")
			outParams = append(outParams, outputParam)
			outParamsResult[outputParam.Key] = output
		}

		printOutParams(ctx, outParamsResult, outParams)

		var judgeObject = &apitestsv2.APITest{}
		if len(exec.Asserts) > 0 {
			var asserts []apistructs.APIAssert
			for _, ass := range exec.Asserts {
				asserts = append(asserts, ass.convert())
			}
			succ, assertResults := judgeObject.JudgeAsserts(outParamsResult, asserts)
			printAssertResults(ctx, succ, assertResults)
			if !succ {
				addNewLine(ctx)
				return fmt.Errorf("sql run Success, but asserts failed")
			}
		}
		for k, v := range outParamsResult {
			meta.OutParamsResult[k] = v
		}
		meta.OutParams = append(meta.OutParams, outParams...)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("db translate commit error %v", err)
	}

	return nil
}

func runSql(execSql string, tx *sql.Tx) (interface{}, error) {
	sqlCmd := strings.ToLower(strings.Split(strings.TrimSpace(execSql), " ")[0])
	switch sqlCmd {
	case "select":
		var results = []map[string]string{}
		result, err := tx.Query(execSql)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logrus.Errorf("sql %v query rollback error %v", execSql, rollbackErr)
			}
			return nil, fmt.Errorf("sql %v query error %v", execSql, err)
		}

		for result.Next() {
			columes, err := result.Columns()
			if err != nil {
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logrus.Errorf("sql %v result columns rollback error %v", execSql, rollbackErr)
				}
				return nil, fmt.Errorf("sql %v result columns error %v", execSql, err)
			}

			var scanObject = make(map[string]string, len(columes))

			err = ScanMap(&scanObject, result)
			if err != nil {
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logrus.Errorf("sql %v result scanMap rollback error %v", execSql, rollbackErr)
				}
				return nil, fmt.Errorf("sql %v result scanMap error %v", execSql, err)
			}
			results = append(results, scanObject)
		}
		return results, nil
	default:
		var results = map[string]interface{}{}

		result, err := tx.Exec(execSql)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logrus.Errorf("sql %v exec rollback error %v", execSql, rollbackErr)
			}
			return nil, fmt.Errorf("sql %v exec error %v", execSql, err)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logrus.Errorf("sql %v exec rowsAffected rollback error %v", execSql, rollbackErr)
			}
			return nil, fmt.Errorf("sql %v exec rowsAffected error %v", execSql, err)
		}

		lastInsertId, err := result.LastInsertId()
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logrus.Errorf("sql %v exec lastInsertId rollback error %v", execSql, rollbackErr)
			}
			return nil, fmt.Errorf("sql %v exec lastInsertId error %v", execSql, err)
		}

		results["lastInsertId"] = lastInsertId
		results["rowsAffected"] = rowsAffected

		return results, nil
	}
}

func genRequest(cfg Conf, bdl *bundle.Bundle) (*Request, error) {
	var req = Request{}

	if len(cfg.DataSource) > 0 {
		mysqlAddon, err := bdl.GetAddon(cfg.DataSource, cfg.OrgID, cfg.UserID)
		if err != nil {
			return nil, fmt.Errorf("getAddonFetchResponseData error %v", err)
		}
		if mysqlAddon == nil {
			return nil, fmt.Errorf("not find this %s mysql service", cfg.DataSource)
		}

		if mysqlAddon.Config[mysqlHost] == nil {
			return nil, fmt.Errorf("addon not find %s", mysqlHost)
		}
		if mysqlAddon.Config[mysqlPort] == nil {
			return nil, fmt.Errorf("addon not find %s", mysqlPort)
		}
		if mysqlAddon.Config[mysqlUsername] == nil {
			return nil, fmt.Errorf("addon not find %s", mysqlUsername)
		}
		if mysqlAddon.Config[mysqlPassword] == nil {
			return nil, fmt.Errorf("addon not find %s", mysqlPassword)
		}

		req.Host = mysqlAddon.Config[mysqlHost].(string)
		req.Port = mysqlAddon.Config[mysqlPort].(string)
		req.User = mysqlAddon.Config[mysqlUsername].(string)
		req.Password = mysqlAddon.Config[mysqlPassword].(string)
	} else {
		req.Host = cfg.Host
		req.Port = cfg.Port
		req.User = cfg.Username
		req.Password = cfg.Password
	}

	if req.Host == "" {
		return nil, fmt.Errorf("%s was empty", mysqlHost)
	}

	if req.Port == "" {
		return nil, fmt.Errorf("%s was empty", mysqlPort)
	}

	if req.User == "" {
		return nil, fmt.Errorf("%s was empty", mysqlUsername)
	}

	if req.Password == "" {
		return nil, fmt.Errorf("%s was empty", mysqlPassword)
	}
	return &req, nil
}

func parseConfFromTask(task *spec.PipelineTask) (Conf, error) {
	envs := mergeEnvs(task)

	var cfg Conf
	if err := envconf.Load(&cfg, envs); err != nil {
		return Conf{}, fmt.Errorf("failed to parse action envs, err: %v", err)
	}
	return cfg, nil
}

func mergeEnvs(task *spec.PipelineTask) map[string]string {
	envs := make(map[string]string)
	for k, v := range task.Extra.PublicEnvs {
		envs[k] = v
	}
	for k, v := range task.Extra.PrivateEnvs {
		envs[k] = v
	}
	return envs
}

type Meta struct {
	AssertResult    bool
	Result          string
	OutParamsResult map[string]interface{}
	Err             error

	OutParams []apistructs.APIOutParam
}

func NewMeta() *Meta {
	return &Meta{
		OutParamsResult: map[string]interface{}{},
	}
}

type KVs []kv
type kv struct {
	k string
	v string
}

func (kvs *KVs) add(k, v string) {
	*kvs = append(*kvs, kv{k, v})
}

func writeMetaFile(ctx context.Context, task *spec.PipelineTask, meta *Meta) {
	log := log_collector.Clog(ctx)

	// kvs 保证顺序
	kvs := &KVs{}

	if meta.Result != "" {
		kvs.add("result", jsonparse.JsonOneLine(ctx, meta.Result))
	}

	if meta.AssertResult {
		kvs.add("assert_result", ResultSuccess)
	} else {
		kvs.add("assert_result", ResultFailed)
	}

	if meta.Err != nil {
		kvs.add("error_info", jsonparse.JsonOneLine(ctx, meta.Err.Error()))
	}

	if len(meta.OutParamsResult) > 0 {
		for _, define := range meta.OutParams {
			v, ok := meta.OutParamsResult[define.Key]
			if !ok {
				continue
			}
			kvs.add(define.Key, jsonparse.JsonOneLine(ctx, v))
		}
	}

	var fields []*apistructs.MetadataField
	for _, kv := range *kvs {
		fields = append(fields, &apistructs.MetadataField{Name: kv.k, Value: kv.v})
	}

	var cb actionagent.Callback
	cb.AppendMetadataFields(fields)
	cb.PipelineID = task.PipelineID
	cb.PipelineTaskID = task.ID

	cbData, _ := json.Marshal(&cb)
	// apitest should ensure that callback to pipeline after doing request
	_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(10*time.Second)).
		Do(func() (bool, error) {
			err := pipelinefunc.CallbackActionFunc(cbData)
			if err != nil {
				log.Errorf("failed to callback, err: %v", err)
				return false, err
			}
			return true, nil
		})
	return
}

func ScanMap(dest interface{}, rows *sql.Rows) error {
	vv := reflect.ValueOf(dest)
	if vv.Kind() != reflect.Ptr || vv.Elem().Kind() != reflect.Map {
		return errors.New("dest should be a map's pointer")
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	newDest := make([]interface{}, len(cols))
	vvv := vv.Elem()

	for i := range cols {
		newDest[i] = reflect.MakeSlice(reflect.SliceOf(vvv.Type().Elem()), 200, 200).Index(0).Addr().Interface()
	}

	err = rows.Scan(newDest...)
	if err != nil {
		return err
	}

	for i, name := range cols {
		vname := reflect.ValueOf(name)
		vvv.SetMapIndex(vname, reflect.ValueOf(newDest[i]).Elem())
	}

	return nil
}
