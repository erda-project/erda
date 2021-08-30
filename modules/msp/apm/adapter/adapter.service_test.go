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

package adapter

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-proto-go/msp/apm/adapter/pb"
)

func Test_adapterService_GetInstrumentationLibrary(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx     context.Context
		request *pb.GetInstrumentationLibraryRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.GetInstrumentationLibraryResponse
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &adapterService{
				p: tt.fields.p,
			}
			got, err := s.GetInstrumentationLibrary(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInstrumentationLibrary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInstrumentationLibrary() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_adapterService_GetInstrumentationLibraryDocs(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx     context.Context
		request *pb.GetInstrumentationLibraryDocsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.GetInstrumentationLibraryDocsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &adapterService{
				p: tt.fields.p,
			}
			got, err := s.GetInstrumentationLibraryDocs(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInstrumentationLibraryDocs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInstrumentationLibraryDocs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
