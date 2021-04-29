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

package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	version       = "v1.0.0"
	genName       = "protoc-gen-go-openapi"
	publishTagKey = "publish"
	privateTagKey = "private"
)

var (
	showVersion = flag.Bool("version", false, "print the version and exit")
	pkgName     *string
	apiPrefix   *string
	customPath  *string
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("%s %v\n", genName, version)
		return
	}

	var flags flag.FlagSet
	pkgName = flags.String("package", "services", "package of output codes")
	apiPrefix = flags.String("prefix", "", "path prefix of API")
	customPath = flags.String("custom", "", "custom path")
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(p *protogen.Plugin) error {
		p.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		var genfiles []*protogen.File
		for _, f := range p.Files {
			if f.Generate {
				genfiles = append(genfiles, f)
			}
		}
		return generateFiles(p, genfiles)
	})
}
