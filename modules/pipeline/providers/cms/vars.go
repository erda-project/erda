package cms

import (
	"github.com/erda-project/erda/modules/pipeline/providers/cms/db"
)

// operations

var (
	DefaultOperationsForKV        = db.DefaultOperationsForKV
	DefaultOperationsForDiceFiles = db.DefaultOperationsForDiceFiles
)

// config type

var (
	ConfigTypeKV       = db.ConfigTypeKV
	ConfigTypeDiceFile = db.ConfigTypeDiceFile
)

type configType string

func (t configType) IsValid() bool {
	return string(t) == ConfigTypeKV || string(t) == ConfigTypeDiceFile
}
