package conf

import _ "embed"

//go:embed openapis/openapis.yaml
var OpenAPIsDefaultConfig string
var OpenAPIsConfigFilePath = "conf/openapis/openapis.yaml"
