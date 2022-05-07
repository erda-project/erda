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

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

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
			Downloadable: true,
			Url:          "https://github.com/erda-project/erda/archive/refs/tags/v2.0.0.tar.gz",
		},
		Readme: []*pb.Readme{
			{
				Lang:   pb.Lang_zh,
				Source: "",
				Text:   "# Create Custom Addon\n\n## Usage\n\n This action have you to create a custom addon.\n",
			},
		},
		Parameters: &pb.Parameters{
			Ins: []string{"params", "outputs"},
			Parameters: []*pb.Parameter{
				{
					Name:        "name",
					In:          "params",
					Description: "custom addon name",
					Required:    true,
					Schema:      &pb.Schema{Type: "string"},
				}, {
					Name:        "tag",
					In:          "params",
					Description: "custom addon tag",
					Required:    true,
					Schema:      &pb.Schema{Type: "string"},
				}, {
					Name:        "success",
					In:          "outputs",
					Description: "create custom addon success or not",
					Required:    true,
					Schema:      &pb.Schema{Type: "string"},
				},
			},
		},
		CreatedAt: timestamppb.Now(),
		Labels:    []string{"official", "addon"},
		Catalog:   "deployment",
	}

	s, err := (&jsonpb.Marshaler{Indent: "  "}).MarshalToString(&p)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)

	var artifact = pb.PutOnProjectArtifactsReq{
		Name:         "erda-create-custom-addon",
		Version:      "1.0",
		Room:         "erda",
		Values:       map[string]string{"example": "..."},
		Presentation: &pb.PresentationRef{Presentation: &p},
		Publisher:    &structpb.Struct{},
		Installer: &pb.Installer{
			Installer: "",
			Spec:      &structpb.Struct{Fields: map[string]*structpb.Value{"releaseID": structpb.NewStringValue("xxx-yyy-zzz")}},
		},
		Syntax:   "1.0",
		RoomKey:  "",
		RoomSign: "",
	}
	s, err = (&jsonpb.Marshaler{Indent: "  "}).MarshalToString(&artifact)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("\n", s)
}
