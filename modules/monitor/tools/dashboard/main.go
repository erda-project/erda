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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Dashboard struct {
	Id         string
	Name       string
	Desc       string
	ViewConfig []interface{}
	Version    string
}

//go:generate echo "Create dashboard sql file starting."
//go:generate go run main.go
//go:generate echo "Create successful, sql file in modules/monitor/tools/dashboard/dist."
func main() {

	pwd, _ := os.Getwd()
	fileName := "dashboard.sql"
	dist := pwd + "/dist"
	dashboardJsonFiles := make([]string, 0, 10)
	err := filepath.Walk(pwd,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".json") {
				dashboardJsonFiles = append(dashboardJsonFiles, path)
				log.Println(path)
			}
			return nil
		})
	if err != nil {
		log.Fatal(err)
	}

	sqls := make([]string, 0)
	for _, filePath := range dashboardJsonFiles {
		file, err := os.Open(filePath)
		defer file.Close()
		if err != nil {
			log.Fatal(err)
		}
		all, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}
		bytes := []byte(string(all))
		if string(all) == "" {
			continue
		}
		var d Dashboard
		err = json.Unmarshal(bytes, &d)
		if err != nil {
			log.Fatal(err)
		}
		stringBytes, err := json.Marshal(d.ViewConfig)
		viewConfig := string(stringBytes)
		viewConfig = strings.ReplaceAll(viewConfig, "'", "\\'")

		sql := fmt.Sprintf("REPLACE INTO `sp_dashboard_block_system` (`id`,`name`,`desc`,`scope`,"+
			"`scope_id`,`version`,`view_config`) VALUES (\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",'%s');",
			d.Id, d.Name, d.Desc, "micro_service", "global", d.Version, viewConfig)
		sqls = append(sqls, sql)
	}

	pathExist, err := pathExists(dist)
	if err != nil {
		return
	}
	if !pathExist {
		err = os.MkdirAll(dist, os.ModePerm)
		err = os.Chmod(dist, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
	fileName = dist + "/" + fileName
	var f *os.File
	if checkFileIsExist(dist + "/" + fileName) {
		f, err = os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC, 0600)
	} else {
		f, err = os.Create(fileName)
	}

	for _, sql := range sqls {
		_, err := f.WriteString(sql)
		_, err = f.WriteString("\n")
		if err != nil {
			f.Close()
			log.Fatal(err)
		}
	}

	defer f.Close()
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
