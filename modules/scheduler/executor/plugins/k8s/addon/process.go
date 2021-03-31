package addon

import (
	"github.com/erda-project/erda/apistructs"
)

func Create(op AddonOperator, sg *apistructs.ServiceGroup) error {
	if err := op.Validate(sg); err != nil {
		return err
	}
	k8syml := op.Convert(sg)

	return op.Create(k8syml)
}

func Inspect(op AddonOperator, sg *apistructs.ServiceGroup) (*apistructs.ServiceGroup, error) {
	if err := op.Validate(sg); err != nil {
		return nil, err
	}
	return op.Inspect(sg)
}

func Remove(op AddonOperator, sg *apistructs.ServiceGroup) error {
	return op.Remove(sg)
}

func Update(op AddonOperator, sg *apistructs.ServiceGroup) error {
	if err := op.Validate(sg); err != nil {
		return err
	}
	k8syml := op.Convert(sg)

	return op.Update(k8syml)
}
