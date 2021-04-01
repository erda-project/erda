package utils

import (
	"errors"

	"github.com/erda-project/erda/apistructs"
)

func CheckAppMode(mode string) error {
	switch mode {
	case string(apistructs.ApplicationModeService), string(apistructs.ApplicationModeLibrary),
		string(apistructs.ApplicationModeBigdata), string(apistructs.ApplicationModeAbility),
		string(apistructs.ApplicationModeMobile), string(apistructs.ApplicationModeApi):
	default:
		return errors.New("invalid request, mode is invalid")
	}
	return nil
}
