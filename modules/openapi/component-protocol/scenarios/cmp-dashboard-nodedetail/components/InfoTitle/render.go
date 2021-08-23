package InfoTitle 

import (
	"context"
	
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func (infoTitle *InfoTitle) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	return nil
}
func RenderCreator() protocol.CompRender {
	return &InfoTitle{}
}
