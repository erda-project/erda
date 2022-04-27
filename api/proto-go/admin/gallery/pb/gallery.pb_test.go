// Copyright (c) 2022 Terminus, Inc.
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

package pb_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/erda-project/erda-infra/pkg/transport/http/encoding/jsonpb"
	"github.com/erda-project/erda-proto-go/admin/gallery/pb"
)

func TestPresentation_MarshalJSON(t *testing.T) {
	var p = pb.Presentation{
		Info: &pb.Info{
			Name:        "create-custom-addon",
			DisplayName: "create-custom-addon",
			Type:        "erda/extension/action",
			Version:     "1.0",
			Summary:     "create a custom addon",
			Description: "create a custom addon",
			Contact: &pb.Contact{
				Name:  "dspo",
				Url:   "https://github.com/dspo",
				Email: "rainchan365@163.com",
			},
			Opensource: &pb.Opensource{
				IsOpenSourced: true,
				Url:           "https://github.com/erda-project/erda-actions",
				License: &pb.License{
					Name: "AGPL",
					Url:  "http://www.gnu.org/licenses/",
				},
			},
			LogoURL: "https://www.erda.cloud/logo",
			Homepage: &pb.Homepage{
				Name:    "Erda Cloud",
				Url:     "https://www.erda.cloud",
				LogoURL: "https://www.erda.cloud/logo",
			},
		},
		Download: &pb.Download{
			Downloadable: false,
		},
		Readme: []*pb.Readme{
			{
				Lang:   pb.Lang_zh,
				Source: "",
				Text:   "# Create Custom Addon\n\n## Usage\n\n This action have you to create a custom addon.\n",
			},
		},
		Parameters: &pb.Parameters{
			Ins:       []string{"params", "outputs"},
			Parameter: nil,
		},
		CreatedAt: timestamppb.Now(),
		Labels:    []string{"official", "addon"},
		Catalog:   "deployment",
	}

	var parameters = []interface{}{
		openapi3.Parameter{
			ExtensionProps: openapi3.ExtensionProps{},
			Name:           "name",
			In:             "params",
			Description:    "custom addon name",
			Required:       true,
			Schema:         &openapi3.SchemaRef{Value: openapi3.NewStringSchema()},
		}, openapi3.Parameter{
			ExtensionProps: openapi3.ExtensionProps{},
			Name:           "tag",
			In:             "params",
			Description:    "custom addon tag",
			Required:       true,
			Schema:         &openapi3.SchemaRef{Value: openapi3.NewStringSchema()},
		}, openapi3.Parameter{
			ExtensionProps: openapi3.ExtensionProps{},
			Name:           "success",
			In:             "outputs",
			Description:    "create custom addon success or not",
			Required:       true,
			Schema:         &openapi3.SchemaRef{Value: openapi3.NewStringSchema()},
		},
	}

	bytes := wrapperspb.Bytes()

	_ = p
	list, err := structpb.NewList(parameters)
	if err != nil {
		t.Fatal(err)
	}
	s, err := (&jsonpb.Marshaler{Indent: "  "}).MarshalToString(list)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)
	//	p.Parameters.Parameter = make([]*structpb.Value, len(parameters))
	//	var orderM = make(map[string]int)
	//	for i, param := range parameters {
	//		orderM[param.In] = i
	//		v, err := structpb.NewValue(param)
	//		if err != nil {
	//			t.Fatal(i, err)
	//		}
	//		p.Parameters.Parameter[i] = v
	//	}
	//
	//	sort.Slice(p.Parameters.Parameter, func(i, j int) bool {
	//		return orderM[parameters[i].In] < orderM[parameters[j].In]
	//	})
	//
	//	s, err := (&jsonpb.Marshaler{Indent: "  "}).MarshalToString(&p)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	t.Log(s)
}
