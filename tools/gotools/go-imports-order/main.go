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
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

const LocalPrefix = "github.com/erda-project/erda"

func checkAllFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip directories like ".git".
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip git ignore files
		relPath, err := filepath.Rel(dir, path)
		if err == nil && ignoreMatcher != nil && ignoreMatcher.MatchesPath(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// Only check Go files.
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Don't check testdata files.
		if strings.Contains("/"+relPath, "/testdata/") {
			return nil
		}

		needsFormat, err := checkFile(dir, path)
		if err != nil {
			return err
		}
		if needsFormat {
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
}

func checkFile(dir, filename string) (bool, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return false, err
	}
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return false, err
	}
	// Don't require headers on generated files.
	if isGenerated(fset, parsed) {
		return false, nil
	}

	if len(parsed.Imports) < 0 {
		return false, nil
	}

	// var stdPkgs,thirdPkgs    []*ast.ImportSpec
	imports := parsed.Imports
	lastGroup, lastPos := -1, 0
	var groups []int
	for _, imp := range imports {
		importPath, _ := strconv.Unquote(imp.Path.Value)
		groupNum := importGroup(LocalPrefix, importPath)
		if groupNum != lastGroup {
			groups = append(groups, groupNum)
			if lastGroup != -1 {
				var lines int
				for _, b := range content[lastPos:imp.Pos()] {
					if b == '\n' {
						lines++
					}
				}
				if lines <= 0 {
					return true, nil
				}
			}
		}
		lastGroup, lastPos = groupNum, int(imp.End())
	}

	// check import order
	lastGroup = -1
	for _, group := range groups {
		if lastGroup > group {
			return true, nil
		}
		lastGroup = group
	}
	return false, nil
}

// importToGroup is a list of functions which map from an import path to a group number.
var importToGroup = []func(localPrefix, importPath string) (num int, ok bool){
	func(localPrefix, importPath string) (num int, ok bool) {
		if localPrefix == "" {
			return
		}
		for _, p := range strings.Split(localPrefix, ",") {
			if strings.HasPrefix(importPath, p) || strings.TrimSuffix(p, "/") == importPath {
				return 3, true
			}
		}
		return
	},
	func(_, importPath string) (num int, ok bool) {
		if strings.HasPrefix(importPath, "appengine") {
			return 2, true
		}
		return
	},
	func(_, importPath string) (num int, ok bool) {
		firstComponent := strings.Split(importPath, "/")[0]
		if strings.Contains(firstComponent, ".") {
			return 1, true
		}
		return
	},
}

func importGroup(localPrefix, importPath string) int {
	for _, fn := range importToGroup {
		if n, ok := fn(localPrefix, importPath); ok {
			return n
		}
	}
	return 0
}

var generatedRx = regexp.MustCompile(`//.*DO NOT EDIT\.?`)

func isGenerated(fset *token.FileSet, file *ast.File) bool {
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			if matched := generatedRx.MatchString(comment.Text); !matched {
				continue
			}
			// Check if comment is at the beginning of the line in source.
			if pos := fset.Position(comment.Slash); pos.Column == 1 {
				return true
			}
		}
	}
	return false
}

var ignoreMatcher *gitignore.GitIgnore

func init() {
	ign, err := gitignore.CompileIgnoreFile(".gitignore")
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("[WARN] .gitignore: ", err)
		}
	}
	ignoreMatcher = ign
}

func main() {
	wd, _ := os.Getwd()
	files, err := checkAllFiles(wd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}
	if len(files) > 0 {
		fmt.Printf("[ERROR] unformatted files (%d):\n", len(files))
		fmt.Println(strings.Join(files, "\n"))
		fmt.Println()
		fmt.Println("Imports order must be:")
		fmt.Println("	1. Golang Standard Package")
		fmt.Println("	2. Third Party Package")
		fmt.Println("	3. Erda Project Package")
		fmt.Println()
		os.Exit(1)
		return
	}
	fmt.Println("Good files !")
}
