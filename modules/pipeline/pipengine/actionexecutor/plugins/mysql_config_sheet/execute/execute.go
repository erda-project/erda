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
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/actionagent"
	"github.com/erda-project/erda/modules/pipeline/pkg/jsonparse"
	"github.com/erda-project/erda/modules/pipeline/pkg/log_collector"
	"github.com/erda-project/erda/modules/pipeline/pkg/pipelinefunc"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/apitestsv2"
	jsonfilter "github.com/erda-project/erda/pkg/encoding/jsonparse"
	"github.com/erda-project/erda/pkg/envconf"
	"github.com/erda-project/erda/pkg/http/httpclient"
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

func Do(ctx context.Context, task *spec.PipelineTask) {
	logger := log_collector.NewLogger().WithContext(context.WithValue(context.Background(), log_collector.CtxKeyCollectorLogID, task.Extra.UUID))
	ctx = context.WithValue(ctx, log_collector.CtxKeyLogger, logger)

	meta := NewMeta()
	var err error

	defer writeMetaFile(ctx, task, meta)
	defer func() {
		if err != nil {
			log_collector.Clog(ctx).Errorf(err.Error())
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

	var req = genRequest(ctx, cfg, meta)
	if req == nil {
		return
	}
	req.ClusterKey = task.Extra.ClusterName

	log_collector.Clog(ctx).Infof("----------- execute sql ----------- ")

	db, err := req.dbOpen()
	if err != nil {
		err = fmt.Errorf("dbOpen error %v", err)
		return
	}
	defer db.Close()

	if cfg.Database != "" {
		log_collector.Clog(ctx).Infof("sql: USE %v", cfg.Database)

		_, err = db.Exec("USE " + cfg.Database)
		if err != nil {
			err = fmt.Errorf("sql %v exec error %v", "USE "+cfg.Database, err)
			return
		}
	}

	log_collector.Clog(ctx).Infof("sql： %v", cfg.Sql)
	runSQL(ctx, cfg, db, meta)
	log_collector.Clog(ctx).Infof("result： %v", meta.Result)

	meta.OutParams = cfg.OutParams
	meta.OutParamsResult = map[string]interface{}{}
	for _, outputParam := range cfg.OutParams {
		output := jsonfilter.FilterJson([]byte(meta.Result), outputParam.Expression, "")
		meta.OutParamsResult[outputParam.Key] = output
	}

	printOutParams(ctx, meta.OutParamsResult, meta)

	var judgeObject = &apitestsv2.APITest{}
	if len(cfg.Asserts) > 0 {

		succ, assertResults := judgeObject.JudgeAsserts(meta.OutParamsResult, cfg.Asserts)
		printAssertResults(ctx, succ, assertResults)
		meta.AssertResult = succ

		if !succ {
			addNewLine(ctx)
			err = fmt.Errorf( "sql run Success, but asserts failed")
			return
		}
	}
	return
}

func runSQL(ctx context.Context, cfg Conf, db *sql.DB, meta *Meta)  {
	switch cfg.SqlType {
	case "query":
		if strings.HasPrefix(cfg.Sql, "select") {
			result, err := db.Query(cfg.Sql)
			if err != nil {
				err = fmt.Errorf("sql %v query error %v", cfg.Sql, err)
				return
			}

			var results = []map[string]interface{}{}
			for result.Next() {
				columes, err := result.Columns()
				if err != nil {
					err = fmt.Errorf("sql %v query error %v", cfg.Sql, err)
					return
				}

				var scanObject =  make([]interface{}, len(columes))
				err = result.Scan(scanObject...)
				if err != nil {
					err = fmt.Errorf("sql %v query error %v", cfg.Sql, err)
					return
				}

				var object = map[string]interface{}{}
				for index, colume := range  columes {
					object[colume] = scanObject[index]
				}
				results = append(results, object)
			}

			meta.Result = jsonparse.JsonOneLine(ctx, results)
			break
		}
		fallthrough
	case "exec":
		result, err := db.Exec(cfg.Sql)
		if err != nil {
			err = fmt.Errorf("sql %v exec error %v", "USE "+cfg.Database, err)
			return
		}
		var results = map[string]interface{}{}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			err = fmt.Errorf("sql %v exec error %v", "USE "+cfg.Database, err)
			return
		}

		lastInsertId, err := result.LastInsertId()
		if err != nil {
			err = fmt.Errorf("sql %v exec error %v", "USE "+cfg.Database, err)
			return
		}

		results["rows_affected"] = rowsAffected
		results["last_Insert_id"] = lastInsertId
		meta.Result = jsonparse.JsonOneLine(ctx, results)
	}
}

func genRequest(ctx context.Context, cfg Conf, meta *Meta) *Request {
	var req = Request{}

	var err error
	defer func() {
		if err != nil {
			log_collector.Clog(ctx).Errorf(err.Error())
			meta.Err = err
			meta.AssertResult = false
		}
	}()

	if len(cfg.DataSource) > 0 {
		mysqlAddon, err := getAddonFetchResponseData(ctx, cfg)
		if err != nil {
			err = fmt.Errorf("getAddonFetchResponseData error %v", err)
		}
		if mysqlAddon == nil {
			err = fmt.Errorf("not find this %s mysql service", cfg.DataSource)
			return nil
		}

		if mysqlAddon.Config[mysqlHost] == nil {
			err = fmt.Errorf("addon not find %s", mysqlHost)
			return nil
		}
		if mysqlAddon.Config[mysqlPort] == nil {
			err = fmt.Errorf("addon not find %s", mysqlPort)
			return nil
		}
		if mysqlAddon.Config[mysqlUsername] == nil {
			err = fmt.Errorf("addon not find %s", mysqlUsername)
			return nil
		}
		if mysqlAddon.Config[mysqlPassword] == nil {
			err = fmt.Errorf("addon not find %s", mysqlPassword)
			return nil
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
		err = fmt.Errorf("%s was empty", mysqlHost)
		return nil
	}

	if req.Port == "" {
		err = fmt.Errorf("%s was empty", mysqlPort)
		return nil
	}

	if req.User == "" {
		err = fmt.Errorf("%s was empty", mysqlUsername)
		return nil
	}

	if req.Password == "" {
		err = fmt.Errorf("%s was empty", mysqlPassword)
		return nil
	}
	return &req
}

func getAddonFetchResponseData(ctx context.Context, cfg Conf) (*apistructs.AddonFetchResponseData, error) {
	var buffer bytes.Buffer
	resp, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(cfg.DiceOpenapiAddr).
		Header("Authorization", cfg.DiceOpenapiToken).
		Path(fmt.Sprintf("/api/addons/%s", cfg.DataSource)).
		Do().Do().Body(&buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to getAddonFetchResponseData, err: %v", err)
	}
	if !resp.IsOK() {
		return nil, fmt.Errorf("failed to getAddonFetchResponseData, statusCode: %d, respBody: %s", resp.StatusCode(), buffer.String())
	}
	var result apistructs.AddonFetchResponse
	respBody := buffer.String()
	if err := json.Unmarshal([]byte(respBody), &result); err != nil {
		return nil, fmt.Errorf("failed to getAddonFetchResponseData, err: %v, json string: %s", err, respBody)
	}
	return &result.Data, nil
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
	Sql             string
	MysqlHost       string
	MysqlUser       string


	AssertResult    bool
	Result string
	OutParamsResult map[string]interface{}
	Err error

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

	if meta.MysqlHost != "" {
		kvs.add("mysql_host", jsonparse.JsonOneLine(ctx, meta.MysqlHost))
	}

	if meta.MysqlUser != "" {
		kvs.add("mysql_user", jsonparse.JsonOneLine(ctx, meta.MysqlUser))
	}

	if meta.Sql != "" {
		kvs.add("sql", jsonparse.JsonOneLine(ctx, meta.Sql))
	}

	if meta.Result != "" {
		kvs.add("result", jsonparse.JsonOneLine(ctx, meta.Result))
	}

	if meta.AssertResult {
		kvs.add("assert_result", ResultSuccess)
	}else{
		kvs.add("assert_result", ResultFailed)
	}

	if meta.Err.Error() != "" {
		kvs.add("error_info", jsonparse.JsonOneLine(ctx, meta.Result))
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
