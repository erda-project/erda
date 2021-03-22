package types

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
)

func TestTypeValues() []apistructs.TestType {
	return []apistructs.TestType{apistructs.UT, apistructs.IT}
}

func TestTypeValueOf(tptype string) (apistructs.TestType, error) {
	switch strings.TrimSpace(tptype) {
	case string(apistructs.UT):
		return apistructs.UT, nil
	case string(apistructs.IT):
		return apistructs.IT, nil
	default:
		return "", errors.Errorf("not supported yet %s", tptype)
	}
}
