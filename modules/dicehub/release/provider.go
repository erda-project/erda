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

package release

import (
	"net/http"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	pb "github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/bundle"
	imagedb "github.com/erda-project/erda/modules/dicehub/image/db"
	"github.com/erda-project/erda/modules/dicehub/release/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

type config struct {
	MaxTimeReserved string `file:"max_time_reserved" env:"RELEASE_MAX_TIME_RESERVED"`
	GCSwitch        bool   `file:"gc_switch" env:"RELEASE_GC_SWITCH"`
}

// +provider
type provider struct {
	Cfg                   *config
	Log                   logs.Logger
	Register              transport.Register `autowired:"service-register" required:"true"`
	DB                    *gorm.DB           `autowired:"mysql-client"`
	Etcd                  *clientv3.Client   `autowired:"etcd"`
	releaseService        *releaseService
	releaseGetDiceService *releaseGetDiceService
	bdl                   *bundle.Bundle
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithScheduler(), bundle.WithCoreServices())

	p.releaseService = &releaseService{
		p:       p,
		db:      &db.ReleaseConfigDB{DB: p.DB},
		imageDB: &imagedb.ImageConfigDB{DB: p.DB},
		bdl:     p.bdl,
		//Etcd:    p.Etcd,
		Config: &releaseConfig{
			MaxTimeReserved: p.Cfg.MaxTimeReserved,
		},
	}
	p.releaseGetDiceService = &releaseGetDiceService{
		p:  p,
		db: &db.ReleaseConfigDB{DB: p.DB},
	}
	if p.Register != nil {
		pb.RegisterReleaseServiceImp(p.Register, p.releaseService, apis.Options(),
			transport.WithHTTPOptions(
				transhttp.WithEncoder(func(rw http.ResponseWriter, r *http.Request, data interface{}) error {
					// TODO because some bug, r.Context() is nil，use r.URL.path temporary
					//info := transport.ContextServiceInfo(r.Context())
					//if info != nil {
					//	if info.Service() == "GetIosPlist" && info.Method() == "GET" {
					//		if resp, ok := data.(*apis.Response); ok && resp != nil {
					//			fmt.Println(reflect.TypeOf(resp.Data))
					//			if dt, ok := resp.Data.(string); ok {
					//				rw.Write([]byte(dt))
					//			}
					//		}
					//	}
					//}
					if strutil.HasPrefixes(r.URL.Path, "/api/releases") && strutil.HasSuffixes(r.URL.Path, "/actions/get-plist") {
						if resp, ok := data.(*apis.Response); ok && resp != nil {
							if dt, ok := resp.Data.(string); ok {
								rw.Write([]byte(dt))
								data = nil
							}
						}
					}
					return encoding.EncodeResponse(rw, r, data)
				}),
			))

		pb.RegisterReleaseGetDiceServiceImp(p.Register, p.releaseGetDiceService, apis.Options(),
			transport.WithHTTPOptions(
				transhttp.WithEncoder(func(rw http.ResponseWriter, r *http.Request, data interface{}) error {
					if resp, ok := data.(*apis.Response); ok && resp != nil {
						if diceYAML, ok := resp.Data.(string); ok {
							if strings.Contains(r.Header.Get("Accept"), "application/x-yaml") {
								rw.Header().Set("Content-Type", "application/x-yaml; charset=utf-8")
								rw.Write([]byte(diceYAML))
							} else { // default: application/json
								yaml, err := diceyml.New([]byte(diceYAML), false)
								if err != nil {
									logrus.Errorf("diceyml new error: %v", err)
									return errors.Errorf("Parse diceyml error.")
								}
								diceJSON, err := yaml.JSON()
								if err != nil {
									logrus.Errorf("diceyml marshal error: %v", err)
									return errors.Errorf("Parse diceyml error.")
								}
								rw.Header().Set("Content-Type", "application/json; charset=utf-8")
								rw.Write([]byte(diceJSON))
							}
							data = nil
						}
					}
					return encoding.EncodeResponse(rw, r, data)
				}),
			))
	}
	// TODO endpoint open GC now. open it after refactor Dicehub
	// Do release Scheduled cleaning tasks
	//if err := p.ReleaseGC(); err != nil {
	//	return err
	//}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.dicehub.release.ReleaseService" || ctx.Type() == pb.ReleaseServiceServerType() || ctx.Type() == pb.ReleaseServiceHandlerType():
		return p.releaseService
	case ctx.Service() == "erda.core.dicehub.release.ReleaseGetDiceService" || ctx.Type() == pb.ReleaseGetDiceServiceServerType() || ctx.Type() == pb.ReleaseGetDiceServiceHandlerType():
		return p.releaseGetDiceService
	}
	return p
}

// ReleaseGC Do release gc Scheduled cleaning tasks
func (p *provider) ReleaseGC() error {
	if p.Cfg.GCSwitch {
		p.ImageGCCron(p.Etcd)
	}
	return nil
}

func init() {
	servicehub.Register("erda.core.dicehub.release", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
