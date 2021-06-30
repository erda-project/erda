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

package pattern

import (
	"io"
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
        },
    },
    "TIME_ZONE": '{{.TimeZone}}',
    "INSTALLED_APPS": ["{{.InstalledApps}}"],
})
apps.populate(django_settings.INSTALLED_APPS)

`

const (
	DjangoMySQLEngine = "django.db.backends.mysql"
	TimeZoneAsiaShanghai = "Asia/Shanghai"
)

type Settings struct {
	Engine        string
	User          string
	Password      string
	Host          string
	Port          int
	Name          string
	TimeZone      string
	InstalledApps string
}

func GenSettings(rw io.ReadWriter, settings Settings) error {
	return generate(rw, "SettingsPattern", SettingsPattern, settings)
}
