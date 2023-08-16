// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"embed"
	"log"
	"os"
	"strings"
)

//go:embed .env
var EnvFile embed.FS

func handleEnvFile() {
	f, err := EnvFile.Open(".env")
	if err != nil {
		log.Fatal("fff", err)
	}
	// scan one by one
	scan := bufio.NewScanner(f)
	scan.Split(bufio.ScanLines)
	for scan.Scan() {
		ss := strings.SplitN(scan.Text(), "=", 2)
		if len(ss) < 2 {
			continue
		}
		if err := os.Setenv(ss[0], ss[1]); err != nil {
			log.Fatal(err)
		}
	}
}
