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

package pygrator

import (
	"io"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

const SettingsPattern = `# encoding: utf8

from django.conf import settings as django_settings
from django.apps import apps


django_settings.configure(**{
    'DATABASES': {
        'default': {
            'ENGINE': '{{.Engine}}',
            "USER": "{{.User}}",
            "PASSWORD": "{{.Password}}",
            "HOST": "{{.Host}}",
            "PORT": {{.Port}},
            'NAME': "{{.Name}}",
            "CHARSET":'utf8mb4,utf8',
        },
    },
    "DEBUG": True,
    "TIME_ZONE": '{{.TimeZone}}',
    "INSTALLED_APPS": ["feature"],
})
apps.populate(django_settings.INSTALLED_APPS)

`

const (
	DjangoMySQLEngine    = "django.db.backends.mysql"
	TimeZoneAsiaShanghai = "Asia/Shanghai"
)

type Settings struct {
	Engine   string
	User     string
	Password string
	Host     string
	Port     int
	Name     string
	TimeZone string
}

func GenSettings(rw io.ReadWriter, settings Settings) error {
	return generate(rw, "SettingsPattern", SettingsPattern, settings)
}

func ParseDSN(dsn string) (*Settings, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ParseDSN")
	}

	var host, portS string
	colon := strings.LastIndexByte(cfg.Addr, ':')
	if colon != -1 && validOptionalPort(cfg.Addr[colon:]) {
		host, portS = cfg.Addr[:colon], cfg.Addr[colon+1:]
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	port, err := strconv.ParseUint(portS, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse port")
	}
	return &Settings{
		Engine:   DjangoMySQLEngine,
		User:     cfg.User,
		Password: cfg.Passwd,
		Host:     host,
		Port:     int(port),
		Name:     cfg.DBName,
		TimeZone: TimeZoneAsiaShanghai,
	}, nil
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}
