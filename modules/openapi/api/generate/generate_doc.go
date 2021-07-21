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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/generate/apistruct"
	"github.com/erda-project/erda/pkg/swagger/oas3"
)

var docJson = make(apistruct.JSON)

var compile = regexp.MustCompile("\\{(.+?)\\}")

func generateDoc(onlyOpenapi bool, resultfile string) {
	var (
		apisM     = make(map[string][]*apis.ApiSpec)
		apisNames = make(map[*apis.ApiSpec]string)
	)

	for i := range APIs {
		api := &APIs[i]
		if api.Host == "" {
			continue
		}
		title := strings.Split(api.Host, ".")[0]
		apisM[title] = append(apisM[title], api)
		apisNames[api] = APINames[i]
	}

	for title, apiList := range apisM {
		var v3 = &openapi3.Swagger{
			ExtensionProps: openapi3.ExtensionProps{},
			OpenAPI:        "3.0.0",
			Components: openapi3.Components{
				Schemas: make(map[string]*openapi3.SchemaRef),
			},
			Info: &openapi3.Info{
				ExtensionProps: openapi3.ExtensionProps{},
				Title:          title,
				Description:    "",
				TermsOfService: "",
				Contact:        nil,
				License:        nil,
				Version:        "default",
			},
			Paths:        make(openapi3.Paths),
			Security:     nil,
			Servers:      nil,
			Tags:         nil,
			ExternalDocs: nil,
		}

		for _, api := range apiList {
			if api.Scheme == "" {
				api.Scheme = "http"
			}

			switch {
			case api.Path == "", // path is invalid
				!strings.HasPrefix(api.Path, "/"), // path do not startswith "/"
				api.Method == "",                  // invalid method
				!api.IsOpenAPI && onlyOpenapi,
				!strings.EqualFold(api.Scheme, "http") && !strings.EqualFold(api.Scheme, "https"): // invalid scheme
				continue
			}

			api.Path = strings.ReplaceAll(api.Path, "<", "{")
			api.Path = strings.ReplaceAll(api.Path, ">", "}")

			var (
				p           *openapi3.PathItem
				refPrefix   = "#/components/schemas/"
				defaultDesc = "write description here"
				requestBody = &openapi3.RequestBodyRef{
					Value: &openapi3.RequestBody{
						Description: defaultDesc,
						Content: map[string]*openapi3.MediaType{
							"application/json": {
								Schema: &openapi3.SchemaRef{
									Ref:   "",
									Value: nil,
								},
								Example:  nil,
								Examples: nil,
								Encoding: nil,
							},
						},
					},
				}
				responseBody = &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: &defaultDesc,
						Content: map[string]*openapi3.MediaType{
							"application/json": {
								Schema: &openapi3.SchemaRef{
									Ref:   "",
									Value: nil,
								},
							},
						},
					},
				}
			)

			if item, ok := v3.Paths[api.Path]; ok {
				p = item
			} else {
				p = new(openapi3.PathItem)

				re := regexp.MustCompile(`{[^/]*}`)
				params := re.FindAllString(api.Path, -1)
				for _, param := range params {
					param := strings.TrimSuffix(strings.TrimPrefix(param, "{"), "}")
					paramItem := openapi3.Parameter{
						Name:        param,
						In:          "path",
						Description: "",
						Style:       "",
						Schema: &openapi3.SchemaRef{
							Ref: "",
							Value: &openapi3.Schema{
								Type:        "string",
								Title:       param,
								Format:      "",
								Description: "",
								Default:     param,
								Example:     param,
							},
						},
					}
					p.Parameters = append(p.Parameters, &openapi3.ParameterRef{Value: &paramItem})
				}
				v3.Paths[api.Path] = p
			}

			var operation = new(openapi3.Operation)
			operation.RequestBody = requestBody
			operation.Responses = openapi3.Responses{"200": responseBody}
			switch strings.ToLower(api.Method) {
			case "connect":
				if p.Connect != nil {
					continue
				}
				p.Connect = operation
			case "delete":
				if p.Delete != nil {
					continue
				}
				p.Delete = operation
			case "get":
				if p.Get != nil {
					continue
				}
				p.Get = operation
				p.Get.RequestBody = nil
			case "head":
				if p.Head != nil {
					continue
				}
				p.Head = operation
				p.Head.RequestBody = nil
			case "options":
				if p.Options != nil {
					continue
				}
				p.Options = operation
			case "patch":
				if p.Patch != nil {
					continue
				}
				p.Patch = operation
			case "post":
				if p.Post != nil {
					continue
				}
				p.Post = operation
			case "put":
				if p.Put != nil {
					continue
				}
				p.Put = operation
			case "trace":
				if p.Trace != nil {
					continue
				}
				p.Trace = operation
			default:
				continue
			}

			operation.Summary = strings.Split(strings.TrimSpace(api.Doc), "\n")[0]
			operation.Description = api.Doc
			tag := api.Group
			if tag == "" {
				tag = "other"
			}
			operation.Tags = append(operation.Tags, tag)
			v3.Tags = append(v3.Tags, &openapi3.Tag{Name: tag})

			if operation.RequestBody != nil {
				if name, schema, err := Struct2OpenapiSchema(api.RequestType); err != nil {
					operation.RequestBody = nil
				} else if name == "" {
					requestBody.Value.Content["application/json"].Schema.Value = schema
				} else if v3.Components.Schemas[name] == nil {
					schemaRef := &openapi3.SchemaRef{Value: schema}
					v3.Components.Schemas[name] = schemaRef
					operation.RequestBody.Value.Content["application/json"].Schema.Ref = refPrefix + name
				} else {
					operation.RequestBody.Value.Content["application/json"].Schema.Ref = refPrefix + name
				}
			}

			if name, schema, err := Struct2OpenapiSchema(api.RequestType); err != nil {
				operation.Responses["200"].Value.Content["application/json"].Schema.Value = &openapi3.Schema{Type: "object"}
			} else if name == "" {
				operation.Responses["200"].Value.Content["application/json"].Schema.Value = schema
			} else if v3.Components.Schemas[name] == nil {
				schemaRef := &openapi3.SchemaRef{Value: schema}
				v3.Components.Schemas[name] = schemaRef
				operation.Responses["200"].Value.Content["application/json"].Schema.Ref = refPrefix + name
			} else {
				operation.Responses["200"].Value.Content["application/json"].Schema.Ref = refPrefix + name
			}
		}

		deDupTags(v3)

		filename := filepath.Base(resultfile)
		filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + ".yml"
		filename = title + "-" + filename
		docf, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		v3Data, err := oas3.MarshalYaml(v3)
		if err != nil {
			panic(err)
		}

		if _, err := docf.Write(v3Data); err != nil {
			panic(err)
		}

		_ = docf.Close()
	}

}

func Struct2OpenapiSchema(i interface{}) (name string, schema *openapi3.Schema, err error) {
	if i == nil {
		return "", nil, errors.New("the input is nil")
	}

	switch i.(type) {
	case time.Time, *time.Time:
		return reflect.TypeOf(i).Name(), openapi3.NewDateTimeSchema(), nil
	}

	switch reflect.ValueOf(i).Type().Kind() {
	case reflect.Bool:
		return reflect.TypeOf(i).Name(), openapi3.NewBoolSchema(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return reflect.TypeOf(i).Name(), openapi3.NewIntegerSchema(), nil

	case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return reflect.TypeOf(i).Name(), openapi3.NewFloat64Schema(), nil

	case reflect.Array, reflect.Slice:
		schema = openapi3.NewArraySchema()
		schema.Items = openapi3.NewSchemaRef("", openapi3.NewStringSchema())
		return reflect.TypeOf(i).Name(), schema, nil

	case reflect.Ptr:
		return Struct2OpenapiSchema(reflect.ValueOf(i).Elem().Interface())

	case reflect.String:
		return reflect.TypeOf(i).Name(), openapi3.NewStringSchema(), nil

	case reflect.Struct:
		typeOf := reflect.TypeOf(i)
		valueOf := reflect.ValueOf(i)
		name = typeOf.Name()
		schema = openapi3.NewObjectSchema()

		for j := 0; j < typeOf.NumField(); j++ {
			if n := typeOf.Field(j).Name[0]; 'a' <= n && n <= 'z' {
				continue
			}
			field := valueOf.Field(j)
			if !field.IsValid() || field.IsZero() {
				var v = openapi3.NewObjectSchema()
				switch field.Kind() {
				case reflect.Bool:
					v = openapi3.NewBoolSchema()
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
					reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					v = openapi3.NewIntegerSchema()
				case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
					v = openapi3.NewFloat64Schema()
				case reflect.Array, reflect.Slice:
					v = openapi3.NewArraySchema()
					v.Items = &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}
				case reflect.String:
					v = openapi3.NewStringSchema()
				}
				schema.Properties[typeOf.Field(j).Name] = &openapi3.SchemaRef{Value: v}
				continue
			}
			if jName, jSchema, err := Struct2OpenapiSchema(field.Interface()); err != nil {
				fmt.Println("Struct2OpenapiSchema error:", err)
				continue
			} else {
				schema.Properties[jName] = &openapi3.SchemaRef{Value: jSchema}
			}

		}
		return name, schema, nil

	default:
		typeOf := reflect.TypeOf(i)
		name = typeOf.Name()
		schema = new(openapi3.Schema)
		schema.Type = "object"
		return name, schema, nil
	}
}

func deDupTags(v3 *openapi3.Swagger) {
	var (
		tagM = make(map[string]*openapi3.Tag)
		tags openapi3.Tags
	)
	for _, tag := range v3.Tags {
		if _, ok := tagM[tag.Name]; ok {
			continue
		}
		tagM[tag.Name] = tag
		tags = append(tags, tag)
	}
	v3.Tags = tags
}
