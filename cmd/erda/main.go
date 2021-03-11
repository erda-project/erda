package main

import (
	"github.com/erda-project/erda-infra/modcom"

	// providers and modules
	_ "github.com/erda-project/erda-infra/providers"
	_ "github.com/erda-project/erda/modules/cmdb"
	_ "github.com/erda-project/erda/modules/openapi"
	_ "github.com/erda-project/erda/modules/pipeline"
	_ "github.com/erda-project/erda/modules/scheduler"
)

func main() {
	modcom.RunWithCfgDir("conf/erda", "erda")
}
