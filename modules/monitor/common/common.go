// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package common

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/common/addon/env"
	"github.com/erda-project/erda/pkg/version"
	"github.com/recallsong/go-utils/config"
	uuid "github.com/satori/go.uuid"
)

var instanceID = uuid.NewV4().String()

// InstanceID .
func InstanceID() string {
	return instanceID
}

// Env .
func Env() {
	config.LoadEnvFile()
}

func loadModuleEnvFile(dir string) {
	path := filepath.Join(dir, ".env."+os.Getenv("DICE_SIZE"))
	config.LoadEnvFileWithPath(path, false)
}

// Hub global variable
var Hub *servicehub.Hub

func prepare() {
	version.PrintIfCommand()
	Env()
	env.Override()
	for _, fn := range initializers {
		fn()
	}
}

var initializers []func()

// RegisterInitializer .
func RegisterInitializer(fn func()) {
	initializers = append(initializers, fn)
}

// Run .
func Run(cfg string) {
	prepare()
	Hub = servicehub.New(servicehub.WithListener(&listener{}))
	Hub.Run("", cfg, os.Args...)
}

// RunWithCfgDir .
func RunWithCfgDir(dir, name string) {
	prepare()
	name = GetEnv("CONF_NAME", name)
	dir = strings.TrimRight(dir, "/")
	os.Setenv("CONF_PATH", dir)
	loadModuleEnvFile(dir)
	cfg := path.Join(dir, name+GetEnv("CONFIG_SUFFIX", ".yml"))
	Hub = servicehub.New(servicehub.WithListener(&listener{}))
	Hub.Run("", cfg, os.Args...)
}

type listener struct{}

func (l *listener) BeforeInitialization(h *servicehub.Hub, config map[string]interface{}) error {
	if _, ok := config["i18n"]; !ok {
		config["i18n"] = nil // i18n is required
	}
	return nil
}

func (l *listener) BeforeExit(h *servicehub.Hub, err error) error {
	return nil
}

func (l *listener) AfterInitialization(h *servicehub.Hub) error {
	api.I18n = h.Service("i18n").(i18n.I18n)
	return nil
}

// GetEnv .
func GetEnv(key, def string) string {
	v := os.Getenv(key)
	if len(v) > 0 {
		return v
	}
	return def
}
