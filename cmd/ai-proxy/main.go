/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	_ "embed"

	"github.com/erda-project/erda-infra/base/servicehub"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy"
	"github.com/erda-project/erda/pkg/common"
)

//go:embed bootstrap.yml
var bootstrap string

func main() {
	common.Run(&servicehub.RunOptions{
		Content: bootstrap,
	})
}
