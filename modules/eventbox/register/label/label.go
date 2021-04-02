// default register labels here
package label

import (
	"github.com/erda-project/erda/modules/eventbox/types"
)

var DefaultLabels = map[types.LabelKey]map[types.LabelKey]interface{}{
	"/TEST_LABEL": {"FAKE": "", "label2": ""},
}
