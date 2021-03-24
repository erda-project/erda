package aop

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
)

// Handle 外部调用统一入口
func Handle(ctx aoptypes.TuneContext) error {
	if !initialized {
		return fmt.Errorf("AOP not initialized")
	}
	typ := ctx.SDK.TuneType
	trigger := ctx.SDK.TuneTrigger
	logrus.Debugf("AOP: type: %s, trigger: %s", typ, trigger)
	chain := tuneGroup.GetTuneChainByTypeAndTrigger(typ, trigger)
	if chain == nil || len(chain) == 0 {
		logrus.Debugf("AOP: type: %s, trigger: %s, tune chain is empty", typ, trigger)
		return nil
	}
	err := chain.Handle(ctx)
	if err != nil {
		logrus.Errorf("AOP: type: %s, trigger: %s, handle failed, err: %v", typ, trigger, err)
		return err
	}
	logrus.Debugf("AOP: type: %s, trigger: %s, handle success", typ, trigger)
	return nil
}
