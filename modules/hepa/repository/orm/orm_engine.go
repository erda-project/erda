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

package orm

import (
	"sync"
	"time"

	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	. "github.com/sirupsen/logrus"
	"github.com/xormplus/core"
	"github.com/xormplus/xorm"
)

type OrmEngineInterface interface {
	xorm.EngineInterface
	SetTableMapper(core.IMapper)
	RegisterSqlMap(xorm.SqlM, ...xorm.Cipher) error
	SqlMapClient(string, ...interface{}) *xorm.Session
}

type EngineMode int8

const (
	SingleMode EngineMode = 1 << iota
	ClusterMode
)

type OrmEngine struct {
	OrmEngineInterface
	mode EngineMode
}

var engine *OrmEngine
var once sync.Once

type SqlLogWrapper struct {
	*Logger
	show bool
}

func (wrapper SqlLogWrapper) ShowSQL(show ...bool) {
	isShow := false
	if len(show) > 0 {
		isShow = show[0]
	}
	wrapper.show = isShow
}

func (wrapper SqlLogWrapper) IsShowSQL() bool {
	return wrapper.show
}

func (wrapper SqlLogWrapper) Level() core.LogLevel {
	switch wrapper.Logger.Level {
	case DebugLevel, TraceLevel:
		return core.LOG_DEBUG
	case InfoLevel:
		return core.LOG_INFO
	case WarnLevel:
		return core.LOG_WARNING
	default:
		return core.LOG_ERR
	}
}

func (wrapper SqlLogWrapper) SetLevel(l core.LogLevel) {
}

var SqlLog SqlLogWrapper

func SetGlobalCacher(engine OrmEngineInterface, expired time.Duration, size int) {
	cacher := xorm.NewLRUCacher2(xorm.NewMemoryStore(), expired, size)
	engine.SetDefaultCacher(cacher)
}

func SetPrefix(engine OrmEngineInterface, prefix string) error {
	if engine == nil {
		return errors.New("engine is nil")
	}
	tbMapper := core.NewPrefixMapper(core.SnakeMapper{}, prefix)
	engine.SetTableMapper(tbMapper)
	return nil
}

func RegisterSqlMap(engine OrmEngineInterface, path string) error {
	if engine == nil {
		return errors.New("engine is nil")
	}

	err := engine.RegisterSqlMap(xorm.Json(path, ".json"))
	if err != nil {
		return errors.Wrap(err, "xorm RegisterSqlMap failed")
	}
	return nil
}

func (engine *OrmEngine) GetEngine() (*xorm.Engine, error) {
	switch engine.mode {
	case SingleMode:
		if raw, ok := engine.OrmEngineInterface.(*xorm.Engine); !ok {
			return nil, errors.New("type downcast to engine failed")
		} else {
			return raw, nil
		}
	case ClusterMode:
		if raw, ok := engine.OrmEngineInterface.(*xorm.EngineGroup); !ok {
			return nil, errors.New("type downcast to engine group failed")
		} else {
			return raw.Engine, nil
		}
	default:
		return nil, errors.New("unkown mode")
	}
}

func NewOrmEngine(driver string, sources []string, options map[string]string) (*OrmEngine, error) {
	if len(sources) == 0 {
		return nil, errors.New("empty source")
	}
	log := &SqlLog
	if len(sources) == 1 {
		innerEngine, err := xorm.NewEngine(driver, sources[0])
		if err != nil {
			return nil, errors.Wrap(err, "xorm NewEngine failed")
		}
		innerEngine.SetLogger(log)
		innerEngine.ShowSQL(true)
		return &OrmEngine{OrmEngineInterface: innerEngine, mode: SingleMode}, nil
	} else {
		optionPolicy := options["policy"]
		var policy xorm.GroupPolicy
		switch optionPolicy {
		case "round_robin":
			policy = xorm.RoundRobinPolicy()
		case "least_conn":
			policy = xorm.LeastConnPolicy()
		case "random":
		default:
			policy = xorm.RandomPolicy()
		}
		innerGroup, err := xorm.NewEngineGroup(driver, sources, policy)
		if err != nil {
			return nil, errors.Wrap(err, "xorm NewEngineGroup failed")
		}
		innerGroup.SetLogger(log)
		innerGroup.ShowSQL(true)
		return &OrmEngine{OrmEngineInterface: innerGroup, mode: ClusterMode}, nil
	}
}

func CreateSingleton(driver string, sources []string, options map[string]string) *OrmEngine {
	once.Do(func() {
		orm_engine, err := NewOrmEngine(driver, sources, options)
		if err != nil {
			panic(err)
		}
		engine = orm_engine
	})
	return engine
}

func GetSingleton() (*OrmEngine, error) {
	if engine == nil {
		Error("ormEngine GetSingleton failed")
		return nil, errors.New("engine is nil")
	}
	return engine, nil
}

func Init() {
	ormEngine := CreateSingleton(config.ServerConf.DbDriver, config.ServerConf.DbSources, map[string]string{})
	err := SetPrefix(ormEngine, config.ServerConf.TableNamePrefix)
	if err != nil {
		panic(err)
	}
	SqlLog = SqlLogWrapper{common.ErrorLog, config.LogConf.ShowSQL}
}
