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

package pattern

import (
	"io"
	"text/template"
)

const EntrypointPattern = `# encoding: utf8

from django.db import connection
import {{.ModuleName}}


if __name__ == "__main__":
    print("Running Erda migration in Python")
    for task in {{.ModuleName}}.entries:
        print("run task: %s.%s" % ({{.ModuleName}}.__name__, task.__name__))
        task()
    [print(query) for query in connection.queries]

`

type Entrypoint struct {
	ModuleName string
}

// GenEntrypoint generates python module entrypoint text and write it to  rw
func GenEntrypoint(rw io.ReadWriter, entrypoint Entrypoint) error {
	return generate(rw, "EntrypointPattern", EntrypointPattern, entrypoint)
}

func generate(rw io.ReadWriter, name, pattern string, data interface{}) error {
	t, err := template.New(name).Parse(pattern)
	if err != nil {
		return err
	}
	if err = t.Execute(rw, data); err != nil {
		return err
	}
	return nil
}