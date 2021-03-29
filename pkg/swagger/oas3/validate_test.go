package oas3_test

//import (
//	"context"
//	"io/ioutil"
//	"testing"
//
//	"github.com/erda-project/erda/pkg/swagger"
//	"github.com/erda-project/erda/pkg/swagger/oas3"
//)
//
//const (
//	testFile  = "./testdata/portal.json"
//	testFile2 = "./testdata/api-design-center.json"
//	testFile3 = "./testdata/gaia-oas3.json"
//)
//
//// go test -v -run TestValidateOAS3
//func TestValidateOAS3(t *testing.T) {
//	data, err := ioutil.ReadFile(testFile2)
//	if err != nil {
//		t.Errorf("failed to ReadFile, err: %v", err)
//	}
//
//	v3, err := swagger.LoadFromData(data)
//	if err != nil {
//		t.Error(err)
//	}
//
//	if err := oas3.ValidateOAS3(context.TODO(), *v3); err != nil {
//		t.Log(err)
//	}
//}
//
//const oas3Text = `
//{
//  "openapi": "3.0.0",
//  "info": {
//    "title": "New API",
//    "description": "# API 设计中心创建的 API 文档。\n\n请在『API 概况』中填写 API 文档的基本信息；在『API列表』新增接口描述；在『数据类型』中定义要引用的数据结构。\n",
//    "version": "default"
//  },
//  "paths": {
//    "/new-resource": {}
//  },
//  "components": {
//    "schemas": {
//      "de": {
//        "type": "object"
//      }
//    }
//  }
//}`
//
//func TestValidateOAS32(t *testing.T) {
//	v3, err := swagger.LoadFromData([]byte(oas3Text))
//	if err != nil {
//		t.Error(err)
//	}
//
//	if err = oas3.ValidateOAS3(context.TODO(), *v3); err != nil {
//		t.Error(err)
//	}
//}
