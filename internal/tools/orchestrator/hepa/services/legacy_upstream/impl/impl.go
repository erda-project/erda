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

package impl

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/xormplus/xorm"

	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common/vars"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/config"
	context1 "github.com/erda-project/erda/internal/tools/orchestrator/hepa/context"
	gw "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/kong"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	db "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/legacy_consumer"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/legacy_upstream"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/micro_api"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/zone"
)

type GatewayUpstreamServiceImpl struct {
	upstreamDb       db.GatewayUpstreamService
	upstreamApiDb    db.GatewayUpstreamApiService
	upstreamRecordDb db.GatewayUpstreamRegisterRecordService
	azDb             db.GatewayAzInfoService
	consumerDb       db.GatewayConsumerService
	consumerBiz      *legacy_consumer.GatewayConsumerService
	apiBiz           *micro_api.GatewayApiService
	zoneBiz          *zone.GatewayZoneService
	kongDb           db.GatewayKongInfoService
	runtimeDb        db.GatewayRuntimeServiceService
	engine           *orm.OrmEngine
	reqCtx           context.Context
}

var once sync.Once

func NewGatewayUpstreamServiceImpl() error {
	once.Do(
		func() {
			upstreamDb, _ := db.NewGatewayUpstreamServiceImpl()
			upstreamApiDb, _ := db.NewGatewayUpstreamApiServiceImpl()
			upstreamRecordDb, _ := db.NewGatewayUpstreamRegisterRecordServiceImpl()
			azDb, _ := db.NewGatewayAzInfoServiceImpl()
			kongDb, _ := db.NewGatewayKongInfoServiceImpl()
			consumerDb, _ := db.NewGatewayConsumerServiceImpl()
			runtimeDb, _ := db.NewGatewayRuntimeServiceServiceImpl()
			engine, _ := orm.GetSingleton()

			legacy_upstream.Service = &GatewayUpstreamServiceImpl{
				upstreamDb:       upstreamDb,
				upstreamApiDb:    upstreamApiDb,
				upstreamRecordDb: upstreamRecordDb,
				azDb:             azDb,
				kongDb:           kongDb,
				consumerDb:       consumerDb,
				consumerBiz:      &legacy_consumer.Service,
				apiBiz:           &micro_api.Service,
				zoneBiz:          &zone.Service,
				engine:           engine,
				runtimeDb:        runtimeDb,
			}
		})
	return nil
}

func (impl GatewayUpstreamServiceImpl) Clone(ctx context.Context) legacy_upstream.GatewayUpstreamService {
	newService := impl
	newService.reqCtx = ctx
	return &newService
}

func (impl GatewayUpstreamServiceImpl) getUpstreamMissingApis(upstreamId string, registerId string) (map[string]orm.GatewayUpstreamApi, error) {
	return impl.getUpstreamApisWithFunc(upstreamId, registerId, impl.upstreamApiDb.SelectInIdsAndDeleted)
}

// getUpstreamApis gets orm.GatewayUpstreamApi from records
func (impl GatewayUpstreamServiceImpl) getUpstreamApis(upstreamId string, registerId string) (map[string]orm.GatewayUpstreamApi, error) {
	return impl.getUpstreamApisWithFunc(upstreamId, registerId, impl.upstreamApiDb.SelectInIds)
}

func (impl GatewayUpstreamServiceImpl) getUpstreamApisWithFunc(upstreamId, registerId string, f func([]string) ([]orm.GatewayUpstreamApi, error)) (map[string]orm.GatewayUpstreamApi, error) {
	res := map[string]orm.GatewayUpstreamApi{}
	record, err := impl.upstreamRecordDb.Get(upstreamId, registerId)
	if err != nil {
		return res, err
	}
	if record == nil {
		return res, errors.Errorf("record of upstreamId[%s] and registerId[%s] not exists", upstreamId, registerId)
	}
	if len(record.UpstreamApis) == 0 {
		return res, nil
	}
	var apiIdList []string
	err = json.Unmarshal(record.UpstreamApis, &apiIdList)
	if err != nil {
		return res, errors.Wrapf(err, "json unmarshal [%s] failed", record.UpstreamApis)
	}
	apiList, err := f(apiIdList)
	if err != nil {
		return res, err
	}
	for _, api := range apiList {
		// 兼容逻辑
		if api.ApiName == "root" {
			res["/"] = api
		} else {
			res[api.ApiName] = api
		}
	}
	return res, nil
}

func safeChange(apis []orm.GatewayUpstreamApi, change func(orm.GatewayUpstreamApi), errorHappened *bool) {
	if *errorHappened {
		return
	}
	size := len(apis)
	batchCount := size / config.ServerConf.RegisterSliceSize
	leaves := size % config.ServerConf.RegisterSliceSize
	log.Debugf("size:%d batchCont:%d leaves:%d", size, batchCount, leaves)
	for i := 0; i < batchCount; i++ {
		for j := 0; j < config.ServerConf.RegisterSliceSize; j++ {
			change(apis[i*config.ServerConf.RegisterSliceSize+j])
			if *errorHappened {
				return
			}
		}
		time.Sleep(time.Duration(config.ServerConf.RegisterInterval) * time.Second)
	}
	for i := size - leaves; i < size; i++ {
		change(apis[i])
		if *errorHappened {
			return
		}
	}
}

func (impl GatewayUpstreamServiceImpl) UpstreamValidAsync(ctx context.Context, context map[string]interface{}, dto *gw.UpstreamRegisterDto) (err error) {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())

	consumerI, ok := context["register_consumer"]
	if !ok {
		err = errors.New("can't find consumer from context")
		return
	}
	consumer, ok := consumerI.(*orm.GatewayConsumer)
	if !ok {
		err = errors.New("acquire consumer failed")
		return
	}
	daoI, ok := context["register_dao"]
	if !ok {
		err = errors.New("can't find dao from context")
		return
	}
	dao, ok := daoI.(*orm.GatewayUpstream)
	if !ok {
		err = errors.New("acquire dao failed")
		return
	}
	return impl.UpstreamValid(ctx, consumer, dao, dto)

}

func (impl GatewayUpstreamServiceImpl) UpstreamValid(ctx context.Context, consumer *orm.GatewayConsumer, dao *orm.GatewayUpstream, dto *gw.UpstreamRegisterDto) (err error) {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())

	prefix := ""
	if dto.PathPrefix != nil {
		prefix = *dto.PathPrefix
	}
	err = impl.upstreamValid(ctx, consumer, dao, dto.RegisterId, prefix, dto.OFlag)
	if err != nil {
		return
	}
	return
}

func (impl GatewayUpstreamServiceImpl) apiNeedUpdate(old, new orm.GatewayUpstreamApi) bool {
	return old.Path != new.Path || old.Method != new.Method || old.Address != new.Address ||
		old.GatewayPath != new.GatewayPath || old.Domains != new.Domains
}

func (impl GatewayUpstreamServiceImpl) upstreamValid(ctx context.Context, consumer *orm.GatewayConsumer, upstream *orm.GatewayUpstream, validId string, aliasPath, oFlag string) error {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())
	l := ctx.(*context1.LogContext).Entry()

	if consumer == nil || upstream == nil || len(upstream.Id) == 0 || validId == "" {
		l.Debugf("consumer:%+v upstream:%+v validId:%s", consumer, upstream, validId) // output for debug
		return errors.New(ERR_INVALID_ARG)
	}
	// upstream db transaction start: get old validId
	session := impl.engine.NewSession()
	defer func() {
		session.Close()
	}()
	err := session.Begin()
	if err != nil {
		return err
	}
	oldValidId, err := impl.upstreamDb.GetValidIdForUpdate(upstream.Id, session)
	if err != nil {
		return err
	}
	err = impl.upstreamDb.UpdateValidId(&orm.GatewayUpstream{
		BaseRow:         orm.BaseRow{Id: upstream.Id},
		ValidRegisterId: validId,
	}, session)
	if err != nil {
		_ = session.Rollback()
		return err
	}
	// extract diff: updates, adds, dels
	var (
		updates, adds, dels []orm.GatewayUpstreamApi
		oldApis             = make(map[string]orm.GatewayUpstreamApi)
		newApis             = make(map[string]orm.GatewayUpstreamApi)
		missingApis         = make(map[string]orm.GatewayUpstreamApi)
		backupApis          = make(map[string]orm.GatewayUpstreamApi)
	)
	l = l.WithField("upstream.id", upstream.Id).
		WithField("oldValidId", oldValidId).
		WithField("validId", validId)
	// get the new api list this time registered if the registerId is new
	// get the old api list last time registered if the registerId is new
	if oldValidId != validId {
		newApis, err = impl.getUpstreamApis(upstream.Id, validId)
		if err != nil {
			return err
		}
		l.WithField("len(oldApis)", len(oldApis)).
			Infoln("oldValidId != validId, get oldApis")
		if oldValidId != "" {
			oldApis, err = impl.getUpstreamApis(upstream.Id, oldValidId)
			if err != nil {
				return err
			}
			l.WithField("len(oldApis)", len(oldApis)).
				Infoln("oldValidId != '' && oldValidId != validId, get oldApis")
			if oFlag == gw.OFlagAppend {
				for k := range oldApis {
					if _, ok := newApis[k]; !ok {
						newApis[k] = oldApis[k]
					}
				}
			}
		}
	}
	// get the apis for the deleted part for the registerId and defer recover them.
	if oldValidId == validId {
		missingApis, err = impl.getUpstreamMissingApis(upstream.Id, validId)
		if err != nil {
			return err
		}
		if oFlag == gw.OFlagAppend {
			newApis, err = impl.getUpstreamApis(upstream.Id, validId)
			if err != nil {
				return err
			}
			for k := range newApis {
				if _, ok := missingApis[k]; ok {
					delete(missingApis, k)
				}
			}
			for k := range missingApis {
				if _, ok := newApis[k]; !ok {
					newApis[k] = missingApis[k]
				}
			}
		} else {
			newApis = missingApis
		}
		l.WithField("len(missingApis)", len(missingApis)).
			Infoln("oldValidId == validId, get the apis for the deleted part for the registerId and defer recover them")
		if len(newApis) == 0 {
			l.Infof("no need to update api, since valid id %s is latest, and no new api no missing api", validId)
			return nil
		}
		defer func() {
			if err == nil {
				for _, api := range missingApis {
					err = impl.upstreamApiDb.Recover(api.Id)
					if err != nil {
						l.Errorf("api recover failed, id:%s", api.Id)
					} else {
						l.Infof("api recover success, id:%s", api.Id)
					}
				}
			}
		}()
	}
	// to relate to a package_api for every upstream_api;
	// if there is not that package_api, add the upstream_api to a list for creating it later.
	// if the upstream_api need to update, add it to a list for updating later.
	for name, upApi := range newApis {
		var oldUpApi orm.GatewayUpstreamApi
		exist := false
		name = strings.TrimSuffix(name, "/")
		if oldUpApi, exist = oldApis[name]; !exist {
			name += "/"
			oldUpApi, exist = oldApis[name]
		}
		if exist {
			upApi.ApiId = oldUpApi.ApiId
			// 兼容老的问题数据
			// 补偿机制：获取最近一次有apiid的记录
			if upApi.ApiId == "" {
				upApi.ApiId = impl.upstreamApiDb.GetLastApiId(&upApi)
			}
			// 二次补偿机制：如果找不到apiid，加入待创建列表, 稍后尝试创建
			if upApi.ApiId == "" {
				adds = append(adds, upApi)
				continue
			}
			if impl.apiNeedUpdate(oldUpApi, upApi) {
				updates = append(updates, upApi)
				backupApis[name] = oldUpApi
			} else {
				err = impl.upstreamApiDb.UpdateApiId(&upApi)
				if err != nil {
					l.WithError(err).Errorf("upstream api update failed!: %+v", upApi)
				}
			}
			delete(oldApis, name)
			continue
		}
		adds = append(adds, upApi)
	}
	// for every old api, if it is related to a package_api, add it to a list for deleting later.
	for _, upApi := range oldApis {
		if upApi.ApiId != "" {
			dels = append(dels, upApi)
		}
	}
	l.Debugf(`upstream valid debug, upstreamId:%s, upstreamName:%s, oldValidId:%s, newValidId:%s, adds:%+v, updates:%+v, dels:%+v`,
		upstream.Id, upstream.UpstreamName, oldValidId, validId, adds, updates, dels)
	l.Infof("upstream valid info, upstreamId:%s, upstreamName:%s, validId:%s, adds:%d, updates:%d, dels:%d",
		upstream.Id, upstream.UpstreamName, validId, len(adds), len(updates), len(dels))
	var added, deled, updated []orm.GatewayUpstreamApi
	errHappened := false
	// for range adds create api, update upstreamApi db
	addFunc := func(upApi orm.GatewayUpstreamApi) {
		if errHappened {
			return
		}
		apiId, err := (*impl.apiBiz).CreateUpstreamBindApi(ctx, consumer, upstream.DiceApp, upstream.DiceService, upstream.RuntimeServiceId, &upApi, aliasPath)
		if err == kong.ErrInvalidReq {
			l.WithError(err).Errorf("invalid api ignored, upApi:%+v", upApi)
			return
		}
		if err != nil {
			l.WithError(err).Errorf("create upstream api failed: %+v, upApi:%+v, exist apiId:%s", err, upApi, apiId)
			errHappened = true
			return
		}
		upApi.ApiId = apiId
		added = append(added, upApi)
		err = impl.upstreamApiDb.UpdateApiId(&upApi)
		if err != nil {
			l.WithError(err).Errorf("upstream api update failed!: %+v", upApi)
			errHappened = true
		}
	}
	safeChange(adds, addFunc, &errHappened)
	// for range updates update api, use kong put method will create if not exist
	updateFunc := func(upApi orm.GatewayUpstreamApi) {
		if errHappened {
			return
		}
		err = (*impl.apiBiz).UpdateUpstreamBindApi(ctx, consumer, upstream.DiceApp, upstream.DiceService, &upApi, aliasPath)
		if err == kong.ErrInvalidReq {
			l.WithError(err).Errorf("invalid api ignored, upApi:%+v", upApi)
			return
		}
		if err != nil {
			l.WithError(err).Errorf("update upstream api failed: %+v, upApi:%+v", err, upApi)
			errHappened = true
			return
		}
		updated = append(updated, upApi)
		err = impl.upstreamApiDb.UpdateApiId(&upApi)
		if err != nil {
			l.WithError(err).Errorf("upstream api update failed!: %+v", upApi)
			errHappened = true
		}
	}
	safeChange(updates, updateFunc, &errHappened)
	// for range dels delete api
	delFunc := func(upApi orm.GatewayUpstreamApi) {
		if errHappened {
			return
		}
		err = (*impl.apiBiz).DeleteUpstreamBindApi(&upApi)
		if err == kong.ErrInvalidReq {
			l.WithError(err).Errorf("invalid api ignored, upApi:%+v", upApi)
			return
		}
		if err != nil {
			l.WithError(err).Errorf("delete upstream api failed: %+v, upApi:%+v", err, upApi)
			errHappened = true
			return
		}
		deled = append(deled, upApi)
	}
	safeChange(dels, delFunc, &errHappened)
	if errHappened {
		// recover:
		for i := len(deled) - 1; i >= 0; i-- {
			upApi := deled[i]
			apiId, err := (*impl.apiBiz).CreateUpstreamBindApi(ctx, consumer, upstream.DiceApp, upstream.DiceService, upstream.RuntimeServiceId, &upApi, aliasPath)
			if err != nil {
				l.WithError(err).Errorf("recover: create upstream api[%+v] failed: %+v", upApi, err)
				continue
			}
			upApi.ApiId = apiId
			err = impl.upstreamApiDb.UpdateApiId(&upApi)
			if err != nil {
				l.WithError(err).Errorf("recover: upstream api[%+v] update failed!: %+v", upApi, err)
			}
			deled = append(deled[:i], deled[i+1:]...)
		}
		for i := len(updated) - 1; i >= 0; i-- {
			upApi := updated[i]
			if backupApi, ok := backupApis[upApi.ApiName]; ok {
				err = (*impl.apiBiz).UpdateUpstreamBindApi(ctx, consumer, upstream.DiceApp, upstream.DiceService, &backupApi, aliasPath)
				if err != nil {
					l.WithError(err).Errorf("recover: update upstream api[%+v] failed: %+v", backupApi, err)
					continue
				}
				err = impl.upstreamApiDb.UpdateApiId(&backupApi)
				if err != nil {
					l.WithError(err).Errorf("recover: upstream api[%+v] update failed!: %+v", backupApi, err)
				}
			} else {
				l.WithError(err).Errorf("recover: can't find backup old api [%+v]", upApi)
			}
			updated = append(updated[:i], updated[i+1:]...)
		}
		for i := len(added) - 1; i >= 0; i-- {
			upApi := added[i]
			err = (*impl.apiBiz).DeleteUpstreamBindApi(&upApi)
			if err != nil {
				l.WithError(err).Errorf("recover: delete upstream api[%+v] failed: %+v", upApi, err)
			}
			added = append(added[:i], added[i+1:]...)
		}
		_ = session.Rollback()
	} else {
		_ = session.Commit()
	}
	l.WithError(err).Infof("upstream valid done, errorHappened:%t upstreamId:%s, upstreamName:%s, validId:%s, added:%d, updated:%d, deled:%d",
		errHappened, upstream.Id, upstream.UpstreamName, validId, len(added), len(updated), len(deled))
	return err
}

func (impl GatewayUpstreamServiceImpl) saveUpstreamApi(session *xorm.Session, dto *gw.UpstreamApiDto, registerId string, upId string) (string, error) {
	if session == nil || dto == nil {
		log.Errorf("invalid session:%+v or dto:%+v", session, dto)
		return "", errors.New(ERR_INVALID_ARG)
	}
	docJson, err := json.Marshal(dto.Doc)
	if err != nil {
		return "", errors.Wrapf(err, "json Marshal [%+v] failed", dto.Doc)
	}
	dao := &orm.GatewayUpstreamApi{
		UpstreamId:  upId,
		RegisterId:  registerId,
		ApiName:     dto.Name,
		Path:        dto.Path,
		GatewayPath: dto.GatewayPath,
		Method:      dto.Method,
		Address:     dto.Address,
		Doc:         docJson,
		Domains:     dto.Domain,
	}
	// deprecated:
	if dto.IsInner {
		dao.IsInner = 1
	}
	return impl.upstreamApiDb.Insert(session, dao)
}

// saveUpstream .
// to create or update the Upstream and record it.
func (impl GatewayUpstreamServiceImpl) saveUpstream(ctx context.Context, dao *orm.GatewayUpstream, dto *gw.UpstreamRegisterDto) (bool, error) {
	apiDtos := dto.ApiList
	registerId := dto.RegisterId
	transSucc := false
	if dao == nil {
		return false, errors.New(ERR_INVALID_ARG)
	}
	session := impl.engine.NewSession()
	defer func(session **xorm.Session) {
		if transSucc {
			_ = (*session).Commit()
		} else {
			_ = (*session).Rollback()
		}
		(*session).Close()
	}(&session)
	err := session.Begin()
	if err != nil {
		return false, errors.WithStack(err)
	}
	_, err = session.Exec("set innodb_lock_wait_timeout=600")
	if err != nil {
		return false, errors.WithStack(err)
	}
	needSave, newCreated, upId, err := impl.upstreamDb.UpdateRegister(ctx, session, dao)
	if err != nil {
		return false, err
	}
	if newCreated {
		// unlock table lock
		_ = session.Commit()
		session.Close()
		session = impl.engine.NewSession()
		err := session.Begin()
		if err != nil {
			return false, errors.WithStack(err)
		}
		// get row lock
		_, _, _, err = impl.upstreamDb.UpdateRegister(ctx, session, dao)
		if err != nil {
			return false, err
		}
	}
	if !needSave {
		return true, nil
	}
	upApiList := []string{}
	for _, apiDto := range apiDtos {
		upApiId, err := impl.saveUpstreamApi(session, &apiDto, registerId, upId)
		if err != nil {
			return false, err
		}
		upApiList = append(upApiList, upApiId)
	}
	apisJson, err := json.Marshal(upApiList)
	if err != nil {
		return false, errors.Wrapf(err, "json marshal [%+v] failed", upApiList)
	}
	err = impl.upstreamRecordDb.Insert(session, &orm.GatewayUpstreamRegisterRecord{
		UpstreamId:   upId,
		RegisterId:   registerId,
		UpstreamApis: apisJson,
	})
	if err != nil {
		return false, err
	}
	transSucc = true
	return transSucc, nil
}

// upstreamRegister .
// retrieves default consumer by orgID, projectID, evn, az,
// if there is no default consumer, creates it.
//
// If the RuntimeName is given, creates runtime service record if not exists, and make relation between the runtime service and the upstream.
//
// Saves the upstream with the request dto.
func (impl GatewayUpstreamServiceImpl) upstreamRegister(ctx context.Context, dto *gw.UpstreamRegisterDto) (*orm.GatewayUpstream, *orm.GatewayConsumer, error) {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())

	var err error
	var az string
	if dto == nil {
		err = errors.Errorf("invalid dto:%+v", dto)
		return nil, nil, err
	}
	if len(dto.Az) == 0 {
		az, err = impl.azDb.GetAz(&orm.GatewayAzInfo{
			Env:       dto.Env,
			OrgId:     dto.OrgId,
			ProjectId: dto.ProjectId,
		})
		if err != nil {
			return nil, nil, err
		}
	} else {
		az = dto.Az
	}
	// upstream db transaction: get register, return if no need update
	dao := &orm.GatewayUpstream{
		OrgId:          dto.OrgId,
		ProjectId:      dto.ProjectId,
		Env:            strings.ToLower(dto.Env),
		Az:             az,
		DiceApp:        dto.AppName,
		DiceService:    dto.DiceService,
		UpstreamName:   dto.UpstreamName,
		LastRegisterId: dto.RegisterId,
	}
	// acquire default consumer
	consumer, err := impl.consumerDb.GetDefaultConsumer(&orm.GatewayConsumer{
		OrgId:     dto.OrgId,
		ProjectId: dto.ProjectId,
		Env:       dto.Env,
		Az:        az,
	})

	if err != nil {
		return nil, nil, err
	}
	if consumer == nil {
		consumer, _, _, err = (*impl.consumerBiz).CreateDefaultConsumer(dto.OrgId, dto.ProjectId, dto.Env, az)
		if err != nil {
			return nil, nil, err
		}
	}
	if dto.RuntimeName != "" {
		session := impl.engine.NewSession()
		_, _ = session.Exec("set innodb_lock_wait_timeout=600")
		err = session.Begin()
		if err != nil {
			session.Close()
			return nil, nil, err
		}
		runtimeService, err := impl.runtimeDb.CreateIfNotExist(session, &orm.GatewayRuntimeService{
			RuntimeName: dto.RuntimeName,
			AppId:       dto.DiceAppId,
			ServiceName: dto.DiceService,
			ProjectId:   dto.ProjectId,
			Workspace:   dto.Env,
			ClusterName: dto.Az,
		})
		if err != nil {
			_ = session.Rollback()
			session.Close()
			return nil, nil, err
		}
		_ = session.Commit()
		session.Close()
		dao.RuntimeServiceId = runtimeService.Id
	}
	_, err = impl.saveUpstream(ctx, dao, dto)
	if err != nil {
		return nil, nil, err
	}
	return dao, consumer, nil
}

func (impl GatewayUpstreamServiceImpl) UpstreamRegisterAsync(ctx context.Context, dto *gw.UpstreamRegisterDto) (result bool, err error) {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())
	l := ctx.(*context1.LogContext).Entry()

	defer func() {
		if err != nil {
			logrus.Errorf("error Happened: %+v", err)
		}
	}()
	dao, consumer, err := impl.upstreamRegister(ctx, dto)
	if dao == nil || consumer == nil {
		err = errors.New("invalid params")
		return
	}
	if dao.AutoBind == 0 {
		log.Infof("dto[%+v] registerd without bind", dto)
		result = true
		return
	}
	meta := make(map[string]interface{})
	meta["register_consumer"] = consumer
	meta["register_dao"] = dao
	go func() {
		if err := impl.UpstreamValidAsync(ctx, meta, dto); err != nil {
			l.WithError(err).Errorln("failed to UpstreamValidAsync")
		}
	}()
	result = true
	return
}

func (impl GatewayUpstreamServiceImpl) UpstreamRegister(ctx context.Context, dto *gw.UpstreamRegisterDto) (result bool, err error) {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())
	l := ctx.(*context1.LogContext).Entry()

	defer func() {
		if err != nil {
			logrus.Errorf("error Happened: %+v", err)
		}
	}()

	// do upstream register
	dao, consumer, err := impl.upstreamRegister(ctx, dto)
	if err != nil {
		err = errors.Wrap(err, "failed to upstreamRegister")
		return
	}
	if dao == nil {
		err = errors.Errorf("invalid params: %s", "invalid GatewayUpstream")
		return
	}
	if consumer == nil {
		err = errors.Errorf("invalid params: %s", "invalid consumer")
		return
	}
	// check autoBind, return if false
	if dao.AutoBind == 0 {
		l.Infof("dto[%+v] registerd without bind", dto)
		result = true
		return
	}
	err = impl.UpstreamValid(ctx, consumer, dao, dto)
	if err != nil {
		return
	}
	result = true
	return
}
