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
	"encoding/json"
	"path/filepath"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
)

// ExpandSchemaRef 展开 SchemaRef
func ExpandSchemaRef(ref *openapi3.SchemaRef, oas3 *openapi3.Swagger) error {

	if ref == nil {
		return errors.New("ExpandSchemaRef is nil")
	}
	if oas3 == nil {
		return errors.New("Swagger is nil")
	}

	schemas := oas3.Components.Schemas // 待引列表
	schemasExists := len(schemas) > 0

	// 如果标准引用
	if ref.Ref != "" {

		// 存在标准引用, 但待引列表为空, 则文档错误
		if !schemasExists {
			return errors.New("Components.Schemas is nil")
		}

		// 在类型列表中查找被引类型
		referencedSchemaRef, ok := schemas[filepath.Base(ref.Ref)]
		if !ok {
			return errors.Errorf("schema referenced not found, reference path: %s", ref.Ref)
		}

		ref.Ref = "" // 删除源 SchemaRef "$ref" 标记所引用的路径

		// 递归展开被引类型
		if err := ExpandSchemaRef(referencedSchemaRef, oas3); err != nil {
			return errors.Wrap(err, "failed to expand referenced ExpandSchemaRef")
		}

		valueCopy(ref, referencedSchemaRef)

		return nil
	}

	if ref.Value == nil {
		return nil
	}

	// 拓展引用合并到 properties
	// 不考虑字段冲突的情况, 字段冲突时, 直接覆盖
	allOf := ref.Value.AllOf
	allOf = append(allOf, ref.Value.AnyOf...)
	allOf = append(allOf, ref.Value.OneOf...)
	if len(allOf) > 0 {
		ref.Value.AllOf = nil
		ref.Value.AnyOf = nil
		ref.Value.OneOf = nil
		for _, of := range allOf {
			if err := ExpandSchemaRef(of, oas3); err != nil {
				return err
			}
			if of.Value == nil || len(of.Value.Properties) == 0 {
				continue
			}
			for propertyName, property := range of.Value.Properties {
				ref.Value.Properties[propertyName] = property
			}
		}
	}

	if properties := ref.Value.Properties; len(properties) > 0 {
		ref.Value.Properties = make(map[string]*openapi3.SchemaRef, 0)
		for _, property := range properties {
			if err := ExpandSchemaRef(property, oas3); err != nil {
				return err
			}
		}
		ref.Value.Properties = properties
	}

	if items := ref.Value.Items; items != nil {
		ref.Value.Items = nil
		if err := ExpandSchemaRef(items, oas3); err != nil {
			return err
		}
		ref.Value.Items = items
	}

	if property := ref.Value.AdditionalProperties; property != nil {
		ref.Value.AdditionalProperties = nil
		if err := ExpandSchemaRef(property, oas3); err != nil {
			return err
		}
		ref.Value.AdditionalProperties = property
	}

	extensions := ref.Value.ExtensionProps.Extensions
	if len(extensions) == 0 {
		return nil
	}
	if len(ref.Value.Properties) == 0 {
		ref.Value.Properties = make(map[string]*openapi3.SchemaRef, 0)
	}

	ref.Value.ExtensionProps.Extensions = nil

	// 如果存在非标准引用 (合并引用)
	for xKey, value := range extensions {
		if matched, _ := regexp.Match("^x-.*-merge$", []byte(xKey)); !matched {
			continue
		}
		data, ok := value.(json.RawMessage)
		if !ok {
			continue
		}

		var ms []map[string]interface{}
		if err := json.Unmarshal(data, &ms); err != nil {
			return errors.Wrapf(err, "failed to Unmarshal %s", xKey)
		}

		for i, m := range ms {
			referencedPath, ok := m["$ref"]
			if !ok {
				continue
			}
			referencedPathStr, ok := referencedPath.(string)
			if !ok {
				continue
			}

			// 存在合并引用, 但待引列表为空, 说明文档错误
			if !schemasExists {
				return errors.New("Components.Schemas is nil")
			}

			// 在类型列表查找被引类型
			referencedSchemaRef, ok := schemas[filepath.Base(referencedPathStr)]
			if !ok {
				return errors.Errorf("schema referenced not found, reference path: %s", referencedPathStr)
			}

			deleteXRef(i, xKey, ref.Value.ExtensionProps.Extensions, ms) // 递归前删除自定义的引用路径, 避免无限递归

			// 递归展开被引类型
			if err := ExpandSchemaRef(referencedSchemaRef, oas3); err != nil {
				return errors.Wrap(err, "failed to expand reference merging ExpandSchemaRef")
			}

			if referencedSchemaRef.Value == nil {
				continue
			}
			if ref.Value == nil {
				ref.Value = referencedSchemaRef.Value
				continue
			}

			// 将被引类型的 required 列表追加到源类型的 required 列表
			ref.Value.Required = append(ref.Value.Required, referencedSchemaRef.Value.Required...)
			ref.Value.Required = dedupeStringSlice(ref.Value.Required)

			if len(referencedSchemaRef.Value.Properties) == 0 {
				referencedSchemaRef.Value.Properties = make(map[string]*openapi3.SchemaRef, 0)
			}

			// 将被引类型的属性追加到源 ExpandSchemaRef
			for propertyName, property := range referencedSchemaRef.Value.Properties {
				ref.Value.Properties[propertyName] = property
			}
		}
	}

	ref.Value.ExtensionProps.Extensions = extensions

	return nil
}

func deleteXRef(i int, key string, extensions map[string]interface{}, ms []map[string]interface{}) {
	if i+1 == len(ms) {
		delete(extensions, key)
	} else {
		data, _ := json.Marshal(ms[i+1:])
		extensions[key] = json.RawMessage(data)
	}
}

// 值拷贝
func valueCopy(dst, src *openapi3.SchemaRef) {
	if src.Value == nil {
		dst.Value = nil
		return
	}

	dst.Value = openapi3.NewSchema()

	srcData, _ := json.Marshal(src.Value)
	if err := json.Unmarshal(srcData, dst.Value); err != nil {
		panic(err)
	}
}
