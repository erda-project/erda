package main

import (
	"embed"
	"fmt"
	"github.com/erda-project/erda-proto-go/msp/apm/expression/pb"
)

//go:embed analyzer
var expressions embed.FS

func main() {
	var e []*pb.Expresion
	readExpression("analyzer", e)
}

func readExpression(name string, e []*pb.Expresion) {
	entries, err := expressions.ReadDir(name)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			readExpression(fmt.Sprintf("%s/%s", name, info.Name()), e)
		}
		fmt.Println(info.Name(), info.Size(), info.IsDir())
	}
}
