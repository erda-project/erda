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

package common

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/base/version"
	"github.com/recallsong/go-utils/config"
	uuid "github.com/satori/go.uuid"
)

var instanceID = uuid.NewV4().String()

// InstanceID .
func InstanceID() string { return instanceID }

// Env .
func Env() {
	config.LoadEnvFile()
}

// GetEnv get environment with default value
func GetEnv(key, def string) string {
	v := os.Getenv(key)
	if len(v) > 0 {
		return v
	}
	return def
}

func loadModuleEnvFile(dir string) {
	path := filepath.Join(dir, ".env")
	config.LoadEnvFileWithPath(path, false)
}

func prepare() {
	version.PrintIfCommand()
	Env()
	for _, fn := range initializers {
		fn()
	}
}

var initializers []func()

// RegisterInitializer .
func RegisterInitializer(fn func()) {
	initializers = append(initializers, fn)
}

var listeners = []servicehub.Listener{}

// RegisterInitializer .
func RegisterHubListener(l servicehub.Listener) {
	listeners = append(listeners, l)
}

// Hub global variable
var Hub *servicehub.Hub

func newHub() *servicehub.Hub {
	var opts []interface{}
	for _, listener := range listeners {
		opts = append(opts, servicehub.WithListener(listener))
	}
	return servicehub.New(opts...)
}

// Run .
func Run(opts *servicehub.RunOptions) {
	prepare()
	opts.Name = GetEnv("CONFIG_NAME", opts.Name)
	cfg := opts.ConfigFile
	if len(cfg) <= 0 && len(opts.Name) > 0 {
		cfg = opts.Name + ".yaml"
	}
	if len(cfg) > 0 {
		suffix := GetEnv("CONFIG_SUFFIX", "")
		if len(suffix) > 0 {
			idx := strings.Index(cfg, ".")
			if idx >= 0 {
				cfg = cfg[:idx]
			}
			cfg = cfg + suffix
			opts.Content = ""
		}
		opts.ConfigFile = cfg

		dir := strings.TrimRight(filepath.Dir(cfg), "/")
		os.Setenv("CONFIG_PATH", dir)
		loadModuleEnvFile(dir)
	}
	if opts.Args == nil {
		opts.Args = os.Args
	}

	// create and run service hub
	Hub := newHub()
	Hub.RunWithOptions(opts)
}
