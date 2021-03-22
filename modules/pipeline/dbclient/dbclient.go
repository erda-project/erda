package dbclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/xormplus/core"
	"github.com/xormplus/xorm"
)

type Client struct {
	*xorm.Engine
}

type Session struct {
	*xorm.Session
	AllowZeroAffectedRows bool
	NeedAutoClose         bool
	NeedNoAutoTime        bool
}

func (client *Client) NewSession(ops ...SessionOption) *Session {
	s := &Session{}
	for _, op := range ops {
		op(s)
	}

	if s.Session == nil {
		s.Session = client.Engine.NewSession()
		s.NeedAutoClose = true
	}

	if s.NeedNoAutoTime {
		s.Session.NoAutoTime()
	}

	return s
}

func (session *Session) Close() {
	if session.NeedAutoClose {
		session.Session.Close()
	}
	return
}

type SessionOption func(*Session)

// WithNoAutoTime 仅作用在当前 session
// 若该 op 后接 WithTxSession 等其他从外部传入 session 的 op，则 WithNoAutoTime 不会在传入的 session 上生效
// 因此需要注意 op 顺序
func WithNoAutoTime() SessionOption {
	return func(session *Session) {
		session.NeedNoAutoTime = true
	}
}
func WithAllowZeroAffectedRows(allow bool) SessionOption {
	return func(session *Session) {
		session.AllowZeroAffectedRows = allow
	}
}
func WithTxSession(_session *xorm.Session) SessionOption {
	return func(session *Session) {
		session.Session = _session
	}
}

var (
	ErrZeroAffectedRows = errors.New("affected rows was 0")
	ErrRecordNotFound   = errors.New("not found")
)

func New() (*Client, error) {
	var cfg clientConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, errors.Wrap(err, "failed to get mysql configuration from env")
	}

	engine, err := xorm.NewEngine("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mysql server")
	}

	engine.SetMapper(core.GonicMapper{})

	engine.ShowSQL(cfg.ShowSQL)
	engine.ShowExecTime(cfg.ShowSQL)

	logLevel := core.LOG_INFO
	if strings.ToUpper(cfg.LogLevel) == "DEBUG" {
		logLevel = core.LOG_DEBUG
	}
	engine.SetLogLevel(logLevel)

	engine.SetMaxOpenConns(cfg.MaxConn)
	engine.SetMaxIdleConns(cfg.MaxIdle)
	engine.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	engine.SetDisableGlobalCache(true)

	return &Client{engine}, nil
}

type clientConfig struct {
	Host            string        `env:"MYSQL_HOST" envDefault:"127.0.0.1"`
	Port            int           `env:"MYSQL_PORT" envDefault:"3306"`
	Username        string        `env:"MYSQL_USERNAME" envDefault:"root"`
	Password        string        `env:"MYSQL_PASSWORD" envDefault:"anywhere"`
	Database        string        `env:"MYSQL_DATABASE" envDefault:"ci"`
	MaxIdle         int           `env:"MYSQL_MAXIDLE" envDefault:"10"`
	MaxConn         int           `env:"MYSQL_MAXCONN" envDefault:"20"`
	ConnMaxLifetime time.Duration `env:"MYSQL_CONNMAXLIFETIME" envDefault:"10s"`
	LogLevel        string        `env:"MYSQL_LOG_LEVEL" envDefault:"INFO"`
	ShowSQL         bool          `env:"MYSQL_SHOW_SQL" envDefault:"false"`
}
