package main

import (
	"github.com/erda-project/erda-infra/modcom"

	// providers and modules
	_ "github.com/erda-project/erda/modules/gittar"
)

func main() {
	modcom.RunWithCfgDir("conf/gittar", "gittar")
}
