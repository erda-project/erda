package main

import (
	"github.com/erda-project/erda-infra/modcom"

	// providers and modules
	_ "github.com/erda-project/erda/modules/ops"
)

func main() {
	modcom.RunWithCfgDir("conf/ops", "ops")
}
