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

package queue

import (
	"context"
	"fmt"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	grpcclient "github.com/erda-project/erda-infra/pkg/transport/grpc"
	"github.com/erda-project/erda-proto-go/core/pipeline/queue/pb"
	"github.com/erda-project/erda/bundle"
)

type config struct {
	DefaultQueueConcurrency int64   `env:"DEFAULT_QUEUE_CONCURRENCY" default:"20"`
	DefaultQueueCPU         float64 `env:"DEFAULT_QUEUE_CPU" default:"30"`
	DefaultQueueMemoryMB    float64 `env:"DEFAULT_QUEUE_MEMORY_MB" default:"30720"` // default: 30GB
}

type provider struct {
	bdl      *bundle.Bundle
	Cfg      *config
	Log      logs.Logger
	Register transport.Register

	QueueManager pb.QueueServiceServer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithErdaServer())
	service := ctx.Service("erda.core.pipeline.queue.QueueService")
	queueService, ok := service.(pb.QueueServiceServer)
	var clientService pb.QueueServiceClient
	clientOK := false
	if conn, connOK := ctx.Service("grpc-client@erda.core.pipeline.queue").(grpcclient.ClientConnInterface); connOK && conn != nil {
		clientService = pb.NewQueueServiceClient(conn)
		clientOK = true
	} else {
		clientService, clientOK = ctx.Service("erda.core.pipeline.queue.QueueService.client").(pb.QueueServiceClient)
	}
	if clientOK && !isNilQueueClient(clientService) {
		queueService = &clientAdapter{client: clientService}
	} else if !ok || isNilQueueService(queueService) {
		return fmt.Errorf("pipeline queue service is unavailable, type=%T", service)
	}
	p.QueueManager = queueService
	return nil
}

func isNilQueueClient(client pb.QueueServiceClient) bool {
	if client == nil {
		return true
	}
	v := reflect.ValueOf(client)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return true
	}
	e := v.Elem()
	if e.Kind() == reflect.Struct {
		field := e.FieldByName("cc")
		if field.IsValid() && (field.Kind() == reflect.Interface || field.Kind() == reflect.Ptr) {
			return field.IsNil()
		}
	}
	return false
}

type clientAdapter struct{ client pb.QueueServiceClient }

func (a *clientAdapter) CreateQueue(ctx context.Context, req *pb.QueueCreateRequest) (*pb.QueueCreateResponse, error) {
	return a.client.CreateQueue(ctx, req)
}
func (a *clientAdapter) GetQueue(ctx context.Context, req *pb.QueueGetRequest) (*pb.QueueGetResponse, error) {
	return a.client.GetQueue(ctx, req)
}
func (a *clientAdapter) PagingQueue(ctx context.Context, req *pb.QueuePagingRequest) (*pb.QueuePagingResponse, error) {
	return a.client.PagingQueue(ctx, req)
}
func (a *clientAdapter) UpdateQueue(ctx context.Context, req *pb.QueueUpdateRequest) (*pb.QueueUpdateResponse, error) {
	return a.client.UpdateQueue(ctx, req)
}
func (a *clientAdapter) DeleteQueue(ctx context.Context, req *pb.QueueDeleteRequest) (*pb.QueueDeleteResponse, error) {
	return a.client.DeleteQueue(ctx, req)
}

func isNilQueueService(service pb.QueueServiceServer) bool {
	if service == nil {
		return true
	}
	v := reflect.ValueOf(service)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func (s *provider) Run(ctx context.Context) error {
	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("queue", &servicehub.Spec{
		Services:     []string{"queue"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: []string{"erda.core.pipeline.queue-client"},
		Description:  "dop queue",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
