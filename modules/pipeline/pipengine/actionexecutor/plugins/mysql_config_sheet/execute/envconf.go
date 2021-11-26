package execute

import (
	"github.com/erda-project/erda/apistructs"
)

// Conf js action param collection
type Conf struct {
	DataSource string `env:"ACTION_DATASOURCE"`
	Host       string `env:"ACTION_HOST"`
	Port       string `env:"ACTION_PORT"`
	Username   string `env:"ACTION_USERNAME"`
	Password   string `env:"ACTION_PASSWORD"`
	Database   string `env:"ACTION_DATABASE"`
	Sql        string `env:"ACTION_SQL" required:"true"`
	SqlType    string `env:"ACTION_SQL_TYPE" default:"exec"`

	OutParams []apistructs.APIOutParam `env:"ACTION_OUT_PARAMS"`
	Asserts   []apistructs.APIAssert   `env:"ACTION_ASSERTS"`

	DiceOpenapiAddr  string `env:"DICE_OPENAPI_ADDR" required:"true"`
	DiceOpenapiToken string `env:"DICE_OPENAPI_TOKEN" required:"true"`
}
