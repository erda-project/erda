package healthcheck

import (
	"context"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/xormplus/xorm"

	"terminus.io/dice/dice/pkg/jsonstore"
)

type ReportMonitor interface {
	Collect() []MonitorCollectMessage
}

type MysqlMonitor struct {
	DbClient *xorm.Engine
	Db       *gorm.DB
}

func (monitor MysqlMonitor) Collect() []MonitorCollectMessage {
	var message []MonitorCollectMessage

	if monitor.DbClient != nil {
		_, err := monitor.DbClient.Exec("select 1")
		if err != nil {
			message = append(message, MonitorCollectMessage{
				Name:    "mysql-connection",
				Status:  Fail,
				Message: "can not connected mysql",
			})
		} else {
			message = append(message, MonitorCollectMessage{
				Name:    "mysql-connection",
				Status:  Ok,
				Message: "success",
			})
		}
	}

	if monitor.Db != nil {
		err := monitor.Db.Exec("select 1").Error
		if err != nil {
			message = append(message, MonitorCollectMessage{
				Name:    "mysql-connection",
				Status:  Fail,
				Message: "can not connected mysql",
			})
		} else {
			message = append(message, MonitorCollectMessage{
				Name:    "mysql-connection",
				Status:  Ok,
				Message: "success",
			})
		}
	}
	return message
}

type RedisMonitor struct {
	RedisCli *redis.Client
}

func (monitor RedisMonitor) Collect() []MonitorCollectMessage {
	var message []MonitorCollectMessage

	if _, err := monitor.RedisCli.Ping().Result(); err != nil {
		message = append(message, MonitorCollectMessage{
			Name:    "redis-connection",
			Status:  Fail,
			Message: "can not connected redis",
		})
	} else {
		message = append(message, MonitorCollectMessage{
			Name:    "redis-connection",
			Status:  Ok,
			Message: "success",
		})

	}
	return message
}

type JsonStoreMonitor struct {
	Store jsonstore.JsonStore
}

func (monitor JsonStoreMonitor) Collect() []MonitorCollectMessage {
	var message []MonitorCollectMessage

	err := monitor.Store.Put(context.Background(), "/ping/value", nil)
	if err != nil {
		message = append(message, MonitorCollectMessage{
			Name:    "jsonStore-connection",
			Status:  Fail,
			Message: "can not connected etcdStore",
		})
	} else {
		message = append(message, MonitorCollectMessage{
			Name:    "jsonStore-connection",
			Status:  Ok,
			Message: "success",
		})
	}

	return message
}
