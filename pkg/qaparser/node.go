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

package qaparser

import (
	"encoding/xml"
)

type XmlNode struct {
	XMLName xml.Name
	Attrs   map[string]string `xml:"-"`
	Content []byte            `xml:",innerxml"`
	Nodes   []XmlNode         `xml:",any"`
}

func (n *XmlNode) Attr(name string) string {
	return n.Attrs[name]
}

func (n *XmlNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type nodeAlias XmlNode
	if err := d.DecodeElement((*nodeAlias)(n), &start); err != nil {
		return err
	}

	content, err := extractContent(n.Content)
	if err != nil {
		return err
	}

	n.Content = content

	n.Attrs = attrMap(start.Attr)
	return nil
}

func attrMap(attrs []xml.Attr) map[string]string {
	if len(attrs) == 0 {
		return nil
	}

	attributes := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		attributes[attr.Name.Local] = attr.Value
	}
	return attributes
}
