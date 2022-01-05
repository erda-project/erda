package custom

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
)

type Model struct {
	StartTime float64 `json:"startTime"`
	EndTime   float64 `json:"endTime"`
	TenantId  string  `json:"tenantId"`
	ServiceId string  `json:"serviceId"`
}

type TransactionInParams struct {
	InParamsPtr *Model
}

func (b *TransactionInParams) CustomInParamsPtr() interface{} {
	if b.InParamsPtr == nil {
		b.InParamsPtr = &Model{}
	}
	return b.InParamsPtr
}

func (b *TransactionInParams) EncodeFromCustomInParams(customInParamsPtr interface{}, stdInParamsPtr *cptype.ExtraMap) {
	cputil.MustObjJSONTransfer(customInParamsPtr, stdInParamsPtr)
}

func (b *TransactionInParams) DecodeToCustomInParams(stdInParamsPtr *cptype.ExtraMap, customInParamsPtr interface{}) {
	cputil.MustObjJSONTransfer(stdInParamsPtr, customInParamsPtr)
}
