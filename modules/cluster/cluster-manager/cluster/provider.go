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

package cluster

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	"github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cluster/cluster-manager/cluster/db"
	"github.com/erda-project/erda/pkg/common/apis"
)

type provider struct {
	Register       transport.Register `autowired:"service-register" required:"true"`
	DB             *gorm.DB           `autowired:"mysql-client"`
	clusterService *ClusterService
	bdl            *bundle.Bundle
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithCoreServices())

	p.clusterService = &ClusterService{
		db:  &db.ClusterDB{DB: p.DB},
		bdl: p.bdl,
	}

	if p.Register != nil {
		pb.RegisterClusterServiceImp(p.Register, p.clusterService, apis.Options(),
			transport.WithHTTPOptions(
				transhttp.WithEncoder(func(rw http.ResponseWriter, r *http.Request, data interface{}) error {
					if resp, ok := data.(*apis.Response); ok && resp != nil {
						switch data := resp.Data.(type) {
						case *pb.ClusterInfo:
							if data.SchedConfig == nil {
								break
							}
							var enableWorkspace *bool
							enableWorkspaceStr := data.SchedConfig.EnableWorkspace
							if enableWorkspaceStr != "" {
								tmp, _ := strconv.ParseBool(enableWorkspaceStr)
								enableWorkspace = &tmp
							}

							m, err := marshal(data)
							if err != nil {
								logrus.Errorf("failed to marshal ClusterInfo, %v", err)
								return err
							}

							sc, _ := m["scheduler"].(map[string]interface{})
							sc["enableWorkspace"] = enableWorkspace
							m["scheduler"] = sc
							resp.Data = m

						case []*pb.ClusterInfo:
							m, err := marshalSlice(data)
							if err != nil {
								logrus.Errorf("failed to marshal ClusterInfo, %v", err)
								return err
							}

							for i, obj := range m {
								schedConfig, ok := obj["scheduler"].(map[string]interface{})
								if !ok {
									continue
								}
								var enableWorspace *bool
								enableWorspaceStr, _ := schedConfig["enableWorkspace"].(string)
								if enableWorspaceStr != "" {
									tmp, _ := strconv.ParseBool(enableWorspaceStr)
									enableWorspace = &tmp
								}
								schedConfig["enableWorkspace"] = enableWorspace
								obj["scheduler"] = schedConfig
								m[i] = obj
							}
							resp.Data = m
						}
					}
					if err := encoding.EncodeResponse(rw, r, data); err != nil {
						logrus.Errorf("failed to encodeResponse, %v", err)
						return err
					}
					return nil
				}),
				transhttp.WithDecoder(func(r *http.Request, out interface{}) error {
					switch out.(type) {
					case *pb.CreateClusterRequest:
						m := make(map[string]interface{})

						var body []byte
						if r.Body != nil {
							body, _ = ioutil.ReadAll(r.Body)
						} else {
							return nil
						}
						if err := json.Unmarshal(body, &m); err != nil {
							logrus.Errorf("failed to unmarshal CreateClusterRequest req body, %v", err)
							return err
						}

						schedConfig, ok := m["scheduler"].(map[string]interface{})
						if !ok {
							logrus.Debugf("Decoder of CreateClusterRequest: empty schedConfig, skip")
							r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
						}

						enableWorkspace, ok := schedConfig["enableWorkspace"].(bool)
						enableWorkspaceStr := ""
						if ok {
							enableWorkspaceStr = strconv.FormatBool(enableWorkspace)
						}

						schedConfig["enableWorkspace"] = enableWorkspaceStr
						m["scheduler"] = schedConfig

						if err := unmarshal(m, out); err != nil {
							return err
						}
						return nil
					case *pb.UpdateClusterRequest:
						m := make(map[string]interface{})

						var body []byte
						if r.Body != nil {
							body, _ = ioutil.ReadAll(r.Body)
						} else {
							return nil
						}
						if err := json.Unmarshal(body, &m); err != nil {
							logrus.Errorf("failed to unmarshal UpdateClusterRequest req body, %v", err)
							return err
						}

						schedConfig, ok := m["scheduler"].(map[string]interface{})
						if !ok {
							logrus.Debugf("Decoder of UpdateClusterRequest: empty schedConfig, skip")
							r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
						}

						enableWorkspace, ok := schedConfig["enableWorkspace"].(bool)
						enableWorkspaceStr := ""
						if ok {
							enableWorkspaceStr = strconv.FormatBool(enableWorkspace)
						}

						schedConfig["enableWorkspace"] = enableWorkspaceStr
						m["scheduler"] = schedConfig

						if err := unmarshal(m, out); err != nil {
							return err
						}
						return nil
					}
					if err := encoding.DecodeRequest(r, out); err != nil {
						logrus.Errorf("failed to decodeRequest, %v", err)
						return err
					}
					return nil
				}),
			))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.clustermanager.ClusterService" || ctx.Type() == pb.ClusterServiceServerType() ||
		ctx.Type() == pb.ClusterServiceHandlerType():
		return p.clusterService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.clustermanager.cluster", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}

func marshal(in interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{})
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func marshalSlice(in interface{}) ([]map[string]interface{}, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	var out []map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func unmarshal(in map[string]interface{}, out interface{}) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, out)
}
