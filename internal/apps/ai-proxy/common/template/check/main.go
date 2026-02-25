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
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("ai-proxy-template-check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	templatesPath := fs.String("path", "cmd/ai-proxy/conf/templates", "path to templates root directory")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	templatesByType, err := template.LoadTemplatesFromFS(nil, os.DirFS(*templatesPath))
	if err != nil {
		fmt.Fprintf(stderr, "template check failed: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, template.TemplateCheckSummary(templatesByType))
	return 0
}
