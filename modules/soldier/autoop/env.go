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

package autoop

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron"
)

// ReadEnv Read the environment variable configuration file of the automation operation and maintenance script
func (a *Action) ReadEnv() error {
	//t, err := template.New("env.conf").Option("missingkey=error").ParseFiles(a.File("env.conf"))
	//if err != nil {
	//	return err
	//}
	//var buf bytes.Buffer
	//err = t.Execute(&buf, a.ClusterInfo)
	//if err != nil {
	//	return err
	//}
	b, err := ioutil.ReadFile(a.File("env.conf"))
	if err != nil {
		return err
	}
	a.Env = make(map[string]string, 10)
	// TODO line number
	//for i, b := 0, buf.Bytes(); i < len(b); i++ {
	for i := 0; i < len(b); i++ {
		switch b[i] {
		case '#':
			for i++; i < len(b); i++ {
				if b[i] == '\n' || b[i] == '\r' {
					break
				}
			}
		case ' ', '\t', '\n', '\r':
		default:
			j := -1
			if isLetter(b[i]) {
				for k := i + 1; k < len(b); k++ {
					if !isLetter(b[k]) && !isNumber(b[k]) {
						if b[k] == '=' {
							j = k
						}
						break
					}
				}
			}
			if j == -1 {
				return fmt.Errorf("invalid key: %d", i)
			}
			key := string(b[i:j])
			var value []byte
			i = j + 1
		LOOP:
			for ; i < len(b); i++ {
				switch b[i] {
				case '\'':
					k := indexSingleQuotes(b[i:])
					if k == -1 {
						return fmt.Errorf("invalid value: %d", j)
					}
					value = append(value, b[i:i+k+1]...)
					i += k
				case '"':
					k := indexDoubleQuotes(b[i:])
					if k == -1 {
						return fmt.Errorf("invalid value: %d", j)
					}
					value = append(value, b[i:i+k+1]...)
					i += k
				case ' ', '\t', '\n', '\r':
					break LOOP
				default:
					value = append(value, b[i])
				}
			}
			a.Env[key] = string(value)
		}
	}
	a.CronTime = trimQuotes(a.Env["CRON_TIME"])
	if a.CronTime != "" {
		a.CronTime = randTime(a.CronTime)
		if _, err := cron.Parse(a.CronTime); err != nil {
			return fmt.Errorf("%s invalid cron time %s: %s", a.Name, a.CronTime, err)
		}
	}
	a.Nodes = trimQuotes(a.Env["NODES"])
	return nil
}

func indexSingleQuotes(b []byte) (i int) {
	i = -1
	if len(b) >= 2 && b[0] == '\'' {
		for j := 1; j < len(b); j++ {
			if b[j] == '\'' {
				i = j
				break
			}
		}
	}
	return
}

func indexDoubleQuotes(b []byte) (i int) {
	i = -1
	if len(b) >= 2 && b[0] == '"' {
		for j := 1; j < len(b); j++ {
			if b[j] == '"' {
				i = j
				break
			} else if b[j] == '\\' {
				j++
			}
		}
	}
	return
}

func isLetter(c byte) bool {
	return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || c == '_'
}

func isNumber(c byte) bool {
	return '0' <= c && c <= '9'
}

func trimQuotes(s string) string {
	if i := len(s) - 1; i > 0 && ((s[0] == '\'' && s[i] == '\'') || (s[0] == '"' && s[i] == '"')) {
		return s[1:i]
	}
	return s
}

func quoting(s string) string {
	var b bytes.Buffer
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '`':
			b.WriteString("\\`")
		case '"':
			b.WriteString(`\"`)
		case '$':
			b.WriteString(`\$`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteByte(s[i])
		}
	}
	b.WriteByte('"')
	return b.String()
}

func randTime(cronTime string) string {
	cronList := strings.Fields(cronTime)
	for index, value := range cronList {
		if value == "?" {
			rand.Seed(time.Now().UnixNano())
			switch index {
			case 0, 1:
				cronList[index] = strconv.Itoa(rand.Intn(60))
			case 2:
				cronList[index] = strconv.Itoa(rand.Intn(24))
			case 4:
				cronList[index] = strconv.Itoa(rand.Intn(12) + 1)
			}
		}
	}
	return strings.Join(cronList, " ")
}
