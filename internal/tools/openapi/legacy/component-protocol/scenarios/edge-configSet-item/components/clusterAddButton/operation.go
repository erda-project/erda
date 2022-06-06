// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clusteraddbutton

import (
	"fmt"

	protocol "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol"
)

type ComponentClusterAddButton struct {
	ctxBundle protocol.ContextBundle
}

func RenderCreator() protocol.CompRender {
	return &ComponentClusterAddButton{}
}
func (c *ComponentClusterAddButton) SetBundle(ctxBundle protocol.ContextBundle) error {
	if ctxBundle.Bdl == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.ctxBundle = ctxBundle
	return nil
}
