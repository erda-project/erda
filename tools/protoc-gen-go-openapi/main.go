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
