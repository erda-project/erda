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

package bundle_test

const requestSchema = `{
                "required": [
                  "approval",
                  "default",
                  "desc",
                  "limits",
                  "name"
                ],
                "type": "object",
                "properties": {
                  "name": {
                    "maxLength": 36,
                    "minLength": 2,
                    "type": "string",
                    "example": "Example"
                  },
                  "desc": {
                    "maxLength": 191,
                    "type": "string",
                    "example": "Example"
                  },
                  "approval": {
                    "type": "string",
                    "example": "auto",
                    "enum": [
                      "auto"
                    ]
                  },
                  "bindDomains": {
                    "type": "array",
                    "items": {
                      "type": "string",
                      "example": "Example"
                    }
                  }
                }
              }`

const responseSchema = `{
        "required": [
          "success"
        ],
        "type": "object",
        "properties": {
          "success": {
            "type": "boolean",
            "example": true
          },
          "err": {
            "required": [
              "code",
              "msg"
            ],
            "type": "object",
            "properties": {
              "code": {
                "type": "string",
                "example": "Example"
              },
              "msg": {
                "type": "string",
                "example": "Example"
              }
            }
          }
        }
      }`

// go test -v -run TestOpenapiSummary
//func TestOpenapiSummary(t *testing.T) {
//	var summaries = []*bundle.APIOperationSummary{{
//		ID:          0,
//		AssetID:     "asset_id",
//		AssetName:   "asset_name",
//		Version:     "jave-demo",
//		Path:        "/some/uniform/resource/identifier",
//		Method:      "GET",
//		OperationID: "retrieve_some_resource",
//	}}
//	data, _ := json.MarshalIndent(summaries, "", "  ")
//	t.Log(string(data))
//}

//func TestAPIOperation(t *testing.T) {
//	requestBody := openapi3.NewObjectSchema()
//	if err := requestBody.UnmarshalJSON([]byte(requestSchema)); err != nil {
//		t.Errorf("failed to UnmarshalJSON requestSchema")
//	}
//	responseBody := openapi3.NewObjectSchema()
//	if err := responseBody.UnmarshalJSON([]byte(responseSchema)); err != nil {
//		t.Errorf("failed to UnmarshalJSON responseBody")
//	}
//
//	oas3.GenExample(requestBody)
//	oas3.GenExample(responseBody)
//
//	var o = bundle.APIOperation{
//		AssetID:     "some_asset",
//		Version:     "java-demo",
//		Path:        "/api/uniform/resource/identifier/{id}",
//		Method:      "GET",
//		Summary:     "查询某某资源的接口",
//		Description: "查询某某资源的接口",
//		OperationID: "查询xx资源",
//		Headers: []*bundle.Parameter{{
//			Name:            "sessionid",
//			Description:     "session id",
//			AllowEmptyValue: false,
//			AllowReserved:   false,
//			Deprecated:      false,
//			Required:        true,
//			Type:            "string",
//			Enum:            nil,
//			Default:         "uuid",
//			Example:         "xxx-xxx-xxx",
//		}},
//		Parameters: []*bundle.Parameter{{
//			Name:            "createAt",
//			Description:     "创建时间",
//			AllowEmptyValue: false,
//			AllowReserved:   false,
//			Deprecated:      false,
//			Required:        false,
//			Type:            "string",
//			Enum:            nil,
//			Default:         time.Now(),
//			Example:         time.Now(),
//		}},
//		RequestBodyDescription: "描述请求体",
//		RequestBodyRequired:    true,
//		RequestBody: []*bundle.RequestBody{{
//			MediaType: "application/json",
//			Body:      requestBody,
//		}},
//		Responses: []*bundle.Response{{
//			StatusCode:  "200",
//			MediaType:   "application/json",
//			Description: "请求成功的响应体",
//			Headers: []*bundle.Parameter{{
//				Name:            "expiredAt",
//				Description:     "过期时间",
//				AllowEmptyValue: false,
//				AllowReserved:   false,
//				Deprecated:      false,
//				Required:        false,
//				Type:            "string",
//				Enum:            nil,
//				Default:         time.Now(),
//				Example:         time.Now(),
//			}},
//			Body: responseBody,
//		}},
//	}
//
//	data, _ := json.MarshalIndent(o, "", "  ")
//	t.Log(string(data), "\n")
//}
