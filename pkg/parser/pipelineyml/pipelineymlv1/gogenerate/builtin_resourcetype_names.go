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
