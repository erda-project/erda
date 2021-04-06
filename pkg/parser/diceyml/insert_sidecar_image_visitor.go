// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package diceyml

import (
	"fmt"

	"github.com/pkg/errors"
)

type InsertSideCarImageVisitor struct {
	DefaultVisitor
	collectError error
	images       map[string]map[string]string
}

func NewInsertSideCarImageVisitor(images map[string]map[string]string) DiceYmlVisitor {
	return &InsertSideCarImageVisitor{images: images}
}

func (o *InsertSideCarImageVisitor) VisitServices(v DiceYmlVisitor, obj *Services) {
	if len(*obj) == 0 {
		return
	}
	o.DefaultVisitor.VisitServices(v, obj)
	if len(o.images) > 0 {
		images := fmt.Sprintf("%v", o.images)
		o.collectError = errors.Wrap(notfoundService, images)
	}
}

func (o *InsertSideCarImageVisitor) VisitService(v DiceYmlVisitor, obj *Service) {
	images, ok := o.images[o.currentService]
	if !ok {
		return
	}
	for name := range obj.SideCars {
		image, ok := images[name]
		if !ok {
			continue
		}
		delete(images, name)
		obj.SideCars[name].Image = image
	}
	if len(images) == 0 {
		delete(o.images, o.currentService)
	}
}

// images: map[servicename]map[sidecarname]image
func InsertSideCarImage(obj *Object, images map[string]map[string]string) error {
	visitor := NewInsertSideCarImageVisitor(images)
	obj.Accept(visitor)
	return visitor.(*InsertSideCarImageVisitor).collectError
}
