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

package format

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda/pkg/readable_time"
	"github.com/erda-project/erda/pkg/strutil"
)

// shortID for containerID
func ShortID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func Time(t interface{}) string {
	var p time.Time
	var err error
	switch v := t.(type) {
	case string:
		p, err = time.Parse(time.RFC3339, v)
		if err != nil {
			p, err = time.Parse("2006-01-02T15:04:05Z0700", v)
			if err != nil {
				return v
			}
		}
	case time.Time:
		p = v
	case *time.Time:
		p = *v
	default:
		return fmt.Sprintf("%v", t)
	}

	return readable_time.Readable(p).String()
}

func ToTimeSpanString(seconds int) string {
	hours := seconds / 3600
	seconds -= hours * 3600
	minutes := seconds / 60
	seconds -= minutes * 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func FormatErrMsg(name string, str string, withHelp bool) string {
	str = strings.TrimSuffix(str, ".") + "."

	if withHelp {
		str += "\nSee '" + os.Args[0] + " " + name + " --help'."
	}

	return fmt.Sprintf("%s: %s\n", os.Args[0], str)
}

func ReadYml(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s", path)
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s", path)
	}
	return content, nil
}

func FormatMemOutput(mem uint64) string {
	if mem/(1024*1024*1024) > 0 {
		return strutil.Concat(fmt.Sprintf("%.2f", float64(mem/(1024*1024))/1024), "GiB")
	}

	return strutil.Concat(strconv.FormatUint(mem/(1024*1024), 10), "MiB")
}
