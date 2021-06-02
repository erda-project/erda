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
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	year      = "2021"
	copyright = fmt.Sprintf(`Copyright (c) %s Terminus, Inc.

This program is free software: you can use, redistribute, and/or modify
it under the terms of the GNU Affero General Public License, version 3
or later ("AGPL"), as published by the Free Software Foundation.

This program is distributed in the hope that it will be useful, but WITHOUT
ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
FITNESS FOR A PARTICULAR PURPOSE.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.`, year)
)

func checkCopyright(dir string) ([]string, error) {
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
		needsCopyright, err := checkFile(dir, path)
		if err != nil {
			return err
		}
		if needsCopyright {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func checkFile(dir, filename string) (bool, error) {
	// Only check Go files.
	if !strings.HasSuffix(filename, ".go") {
		return false, nil
	}
	// Don't check testdata files.
	normalized := strings.TrimPrefix(filepath.ToSlash(filename), filepath.ToSlash(dir))
	if strings.Contains(normalized, "/testdata/") {
		return false, nil
	}
	// Don't check ingore files.
	if isIgnoreFile(normalized) {
		return false, nil
	}

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
	shouldAddCopyright := true
	for _, c := range parsed.Comments {
		// The copyright should appear before the package declaration.
		if c.Pos() > parsed.Package {
			break
		}

		if strings.HasPrefix(c.Text(), copyright) {
			shouldAddCopyright = false
			break
		}
	}
	return shouldAddCopyright, nil
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

func loadIgnoreFiles(dir string) []string {
	cfgfile := filepath.Join(dir, ".licenserc.json")
	byts, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		return nil
	}
	var config struct {
		Ignore []string `json:"ignore"`
	}
	err = json.Unmarshal(byts, &config)
	if err != nil {
		fmt.Println(fmt.Errorf("fail to parse %s", cfgfile))
		os.Exit(1)
	}
	return config.Ignore
}

var ignoreFiles []string

func isIgnoreFile(filename string) bool {
	filename = strings.TrimLeft(filename, "/")
	for _, file := range ignoreFiles {
		file = strings.TrimLeft(file, "/")
		if strings.HasPrefix(filename, file) {
			return true
		}
	}
	return false
}

func init() {
	wd, _ := os.Getwd()
	ignoreFiles = loadIgnoreFiles(wd)
}

func main() {
	wd, _ := os.Getwd()
	files, err := checkCopyright(wd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}
	if len(files) > 0 {
		fmt.Println("invalid copyright files:")
		fmt.Println(strings.Join(files, "\n"))
		os.Exit(1)
		return
	}
	fmt.Println("Good files !")
}
