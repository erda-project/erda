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

//import (
//	"fmt"
//	"go/ast"
//	"go/parser"
//	"go/token"
//	"io"
//	"os"
//)
//
//const (
//	srcFile = "../builtin_resourcetype.go"
//	dstFile = "../builtin_resourcetype_names.go"
//)
//
//func main() {
//	w, err := os.OpenFile(dstFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
//	defer func() {
//		if r := recover(); r != nil {
//			fmt.Printf("failed to generate %s: %v", dstFile, r)
//			os.Remove(dstFile)
//		}
//	}()
//	if err != nil {
//		panic(err)
//	}
//
//	var builtinResTypeNames []string
//
//	const builtinResTypeStr = "builtinResType"
//
//	fset := token.NewFileSet()
//	f, err := parser.ParseFile(fset, srcFile, nil, 0)
//	if err != nil {
//		panic(err)
//	}
//	for _, decl := range f.Decls {
//		gen, ok := decl.(*ast.GenDecl)
//		if ok && gen.Tok == token.CONST {
//			for _, spec := range gen.Specs {
//				v, ok := spec.(*ast.ValueSpec)
//				if ok {
//					for _, ident := range v.Names {
//						typeIdent, ok := v.Type.(*ast.Ident)
//						if ok && typeIdent.Name == builtinResTypeStr {
//							builtinResTypeNames = append(builtinResTypeNames, ident.Name)
//						}
//					}
//				}
//			}
//		}
//	}
//	fmt.Printf("BuiltinResTypeNames: %+v\n", builtinResTypeNames)
//
//	io.WriteString(w, `// DO NOT EDIT!!!
//// See go:generate at github.com/erda-project/erda/modules/pipeline/pipelineyml/builtin_resourcetype.go
//
//package pipelineymlv1
//
//var BuiltinResTypeNames = []string{
//`)
//
//	for _, name := range builtinResTypeNames {
//		io.WriteString(w, "\tstring("+name+"),\n")
//	}
//	io.WriteString(w, "}\n")
//}
