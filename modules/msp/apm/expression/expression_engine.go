package expression

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda-proto-go/msp/apm/expression/pb"
	"strings"
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
