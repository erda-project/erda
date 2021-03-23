// post_check.go 定义了 Load 之后需要进行的检查

package conf

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/strutil"
)

func checkActionTypeMapping(cfg *Conf) {
	cfg.ActionTypeMapping = make(map[string]string)
	// 可以 环境变量 中临时覆盖映射关系
	for _, v := range strutil.Split(cfg.ActionTypeMappingStr, ",", true) {
		vv := strutil.Split(v, ":", true)
		if len(vv) != 2 {
			logrus.Errorf("[alert] invalid action type mapping: %q", v)
			continue
		}
		cfg.ActionTypeMapping[vv[0]] = vv[1]
	}
}
