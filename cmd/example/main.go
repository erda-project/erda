package main

import (
	"github.com/erda-project/erda-infra/modcom"
	_ "github.com/erda-project/erda-infra/providers"
	_ "github.com/erda-project/erda/modules/example"
)

func main() {
	modcom.RunWithCfgDir("conf/example", "example")
}
