package main

import (
	"github.com/erda-project/erda-infra/modcom"

	// providers and modules
	_ "github.com/erda-project/erda/modules/pipeline"
)

func main() {
	modcom.RunWithCfgDir("conf/pipeline", "pipeline")
}
