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

package expression

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erda-project/erda-proto-go/msp/apm/expression/pb"
)

const (
	Analyzer = "analyzer"
	Alert    = "alert"
)

var Types = map[string]embed.FS{
	Analyzer: analyzerExpressions,
	Alert:    alertExpressions,
}

func GetFS(name string) embed.FS {
	return Types[name]
}

//go:embed analyzer
var analyzerExpressions embed.FS

//go:embed alert
var alertExpressions embed.FS

func readExpression(fs embed.FS, name string, list *[]*pb.Expression) {
	entries, err := fs.ReadDir(name)
	if err != nil {
		fmt.Printf("Read dir(%s) with error: %+v\n", name, err)
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			fmt.Printf("Read fs entry with error: %+v\n", err)
		}

		filename := fmt.Sprintf("%s/%s", name, info.Name())
		if info.IsDir() {
			readExpression(fs, filename, list)
		}

		if !strings.HasSuffix(info.Name(), ".json") {
			continue
		}

		file, err := fs.ReadFile(filename)
		if err != nil {
			fmt.Printf("Read file(%s) with error: %+v\n", filename, err)
			continue
		}
		var expression pb.Expression
		err = json.Unmarshal(file, &expression)
		if err != nil {
			fmt.Printf("Unmarshal file(%s) with error: %+v\n", filename, err)
			continue
		}
		*list = append(*list, &expression)
	}
}
