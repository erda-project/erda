package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/publicsuffix"

	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/apitest/logic/cookiejar"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/apitestsv2"
	"github.com/erda-project/erda/pkg/envconf"
)

const CookieJar = "cookieJar"

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

func parseConfFromTask(task *spec.PipelineTask) (EnvConfig, error) {
	envs := mergeEnvs(task)

	var cfg EnvConfig
	if err := envconf.Load(&cfg, envs); err != nil {
		return EnvConfig{}, fmt.Errorf("failed to parse action envs, err: %v", err)
	}
	return cfg, nil
}

// Do do api test.
//
// attention:
//   push log to collector
//   save metadata
func Do(ctx context.Context, task *spec.PipelineTask) {
	// print logo
	printLogo()

	// parse conf from task
	cfg, err := parseConfFromTask(task)
	if err != nil {
		log.Errorf("failed to parse config from task, err: %v", err)
		return
	}

	// get api info and pretty print
	apiInfo := generateAPIInfoFromEnv(cfg)

	printOriginalAPIInfo(apiInfo)

	// success
	var success = true
	defer func() {
		if !success {
			return
		}
	}()

	// defer create metafile
	meta := NewMeta()
	meta.OutParamsDefine = apiInfo.OutParams
	defer writeMetaFile(task, meta)

	// global config
	var apiTestEnvData *apistructs.APITestEnvData
	caseParams := make(map[string]*apistructs.CaseParams)
	if cfg.GlobalConfig != nil {
		apiTestEnvData = &apistructs.APITestEnvData{}
		apiTestEnvData.Domain = cfg.GlobalConfig.Domain
		apiTestEnvData.Header = cfg.GlobalConfig.Header
		apiTestEnvData.Global = make(map[string]*apistructs.APITestEnvVariable)
		for name, item := range cfg.GlobalConfig.Global {
			apiTestEnvData.Global[name] = &apistructs.APITestEnvVariable{
				Value: item.Value,
				Type:  item.Type,
			}
			caseParams[name] = &apistructs.CaseParams{
				Key:   name,
				Type:  item.Type,
				Value: item.Value,
			}
		}
	}

	// add cookie jar
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if apiTestEnvData != nil && apiTestEnvData.Header != nil && len(apiTestEnvData.Header[CookieJar]) > 0 {
		var cookies cookiejar.Cookies
		err := json.Unmarshal([]byte(apiTestEnvData.Header[CookieJar]), &cookies)
		if err != nil {
			success = false
			log.Errorf("failed to unmarshal cookieJar from header, err: %v\n", err)
			return
		}
		cookieJar.SetEntries(cookies)
	}
	hc := http.Client{Jar: cookieJar}
	printGlobalAPIConfig(apiTestEnvData)

	// do apiTest
	apiTest := apitestsv2.New(apiInfo)
	apiReq, apiResp, err := apiTest.Invoke(&hc, apiTestEnvData, caseParams)
	printRenderedHTTPReq(apiReq)
	meta.Req = apiReq
	meta.Resp = apiResp
	meta.CookieJar = cookieJar.GetEntries()
	if apiResp != nil {
		printHTTPResp(apiResp)
	}
	if err != nil {
		meta.Result = ResultFailed
		log.Errorf("failed to do api test, err: %v", err)
		success = false
		return
	}

	// outParams store in metafile for latter use
	outParams := apiTest.ParseOutParams(apiTest.API.OutParams, apiResp, caseParams)
	printOutParams(outParams, meta)

	// judge asserts
	if len(apiTest.API.Asserts) > 0 {
		// 目前有且只有一组 asserts
		for _, group := range apiTest.API.Asserts {
			succ, assertResults := apiTest.JudgeAsserts(outParams, group)
			printAssertResults(succ, assertResults)
			if !succ {
				addNewLine()
				log.Errorf("API Test Success, but asserts failed")
				success = false
				return
			}
		}
	}

	meta.Result = ResultSuccess

	addNewLine(2)
	log.Println("API Test Success")
}
