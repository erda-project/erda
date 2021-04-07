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

package oas3

import (
	"bytes"
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/yaml"
)

func MarshalJson(v3 *openapi3.Swagger) ([]byte, error) {
	return json.Marshal(v3)
}

func MarshalJsonIndent(v3 *openapi3.Swagger, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v3, prefix, indent)
}

func MarshalYaml(v3 *openapi3.Swagger) ([]byte, error) {
	var buf = bytes.NewBuffer(nil)
	openapi, _ := yaml.Marshal(v3.OpenAPI)
	buf.WriteString("openapi: ")
	buf.Write(openapi)

	info, _ := yaml.Marshal(v3.Info)
	buf.WriteString("\ninfo:")
	buf.Write(indent(info))
	if len(v3.Servers) > 0 {
		servers, _ := yaml.Marshal(v3.Servers)
		buf.WriteString("\nservers:")
		buf.Write(indent(servers))
	}

	paths, _ := yaml.Marshal(v3.Paths)
	buf.WriteString("\npaths:")
	buf.Write(indent(paths))

	components, _ := yaml.Marshal(v3.Components)
	buf.WriteString("\ncomponents:")
	buf.Write(indent(components))

	if len(v3.Security) > 0 {
		security, _ := yaml.Marshal(v3.Security)
		buf.WriteString("\nsecurity:")
		buf.Write(indent(security))
	}

	if len(v3.Tags) > 0 {
		tags, _ := yaml.Marshal(v3.Tags)
		buf.WriteString("\ntags:")
		buf.Write(indent(tags))
	}

	if v3.ExternalDocs != nil {
		externalDocs, _ := yaml.Marshal(v3.ExternalDocs)
		buf.WriteString("\nexternalDocs:")
		buf.Write(indent(externalDocs))
	}

	if len(v3.Extensions) > 0 {
		extensions, _ := yaml.Marshal(v3.Extensions)
		buf.WriteString("\n")
		buf.Write(extensions)
	}

	return buf.Bytes(), nil
}

func indent(s []byte) []byte {
	split := bytes.Split(s, []byte{'\n'})
	var buf = bytes.NewBuffer(nil)
	for _, ele := range split {
		buf.WriteString("\n")
		buf.WriteString("  ")
		buf.Write(ele)
	}
	return buf.Bytes()
}
