package apitestsv2

import (
	"fmt"
	"reflect"

	"github.com/erda-project/erda/apistructs"
)

func checkBodyType(body apistructs.APIBody, expectType reflect.Kind) error {
	if body.Content == nil {
		return nil
	}
	_type := reflect.TypeOf(body.Content).Kind()
	if _type != expectType {
		return fmt.Errorf("invalid body type (%s) while Content-Type is %s", _type.String(), body.Type.String())
	}
	return nil
}
