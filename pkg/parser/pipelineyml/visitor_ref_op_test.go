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

package pipelineyml

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/strutil"
)

func TestOutputRegexp(t *testing.T) {
	reg := regexp.MustCompile(`\${([^:]+):OUTPUT:([^:]+)}`)

	assert.True(t, reg.MatchString("${git-checkout:OUTPUT:image}"))
	assert.True(t, reg.MatchString("${java:OUTPUT:image}"))
	assert.True(t, reg.MatchString("${release:OUTPUT:releaseID}"))
	assert.True(t, reg.MatchString("${中文:OUTPUT:releaseID}"))

	assert.False(t, reg.MatchString("${release:OUTPUT1:releaseID}"))
	assert.False(t, reg.MatchString("${:OUTPUT:releaseID}"))
	assert.False(t, reg.MatchString("${release:OUTPUT:}"))

	spew.Dump(reg.FindAllStringSubmatch(`
${release:OUTPUT:releaseID}
${java:OUTPUT:image}
${中文:OUTPUT:image}
${~:OUTPUT:image}
${{:OUTPUT:image}
`, -1))
}

func TestAllRef(t *testing.T) {
	reg := regexp.MustCompile(`\${([^:]+):([^:]+):(.+)}`)

	fmt.Println(strutil.ReplaceAllStringSubmatchFunc(reg, "${git-checkout:OUTPUT:image:xxx:xxx}", func(sub []string) string {
		fmt.Println(sub[0])
		alias := sub[1]
		op := sub[2]
		key := sub[3]
		return alias + "::" + op + "::" + key
	}))

	assert.True(t, reg.MatchString("${git-checkout:OUTPUT:image}"))
	assert.True(t, reg.MatchString("${git-checkout:OUTPUT:image:hello}"))
	assert.True(t, reg.MatchString("${git-check{out:OUTPUT:image:hello}"))
}

//func TestRef(t *testing.T) {
//	re := regexp.MustCompile(`\$\{\{.*\\}\\}`)
//
//	s := `${{java}}`
//	allMatch := re.FindAllString(s, -1)
//	assert.True(t, len(allMatch) == 1)
//	assert.Equal(t, s, allMatch[0])
//
//	s = `${HOME1:-xxx}`
//	allMatch = re.FindAllString(s, -1)
//	assert.True(t, len(allMatch) == 1)
//	assert.Equal(t, s, allMatch[0])
//
//	s = `${java:OUTPUT:image}`
//	allMatch = re.FindAllString(s, -1)
//	assert.True(t, len(allMatch) == 1)
//	assert.Equal(t, s, allMatch[0])
//
//	s1 := `${java}`
//	s2 := `${java:OUTPUT:image}`
//	s = strutil.Concat(s1, " ", s2)
//	allMatch = re.FindAllString(s, -1)
//	assert.True(t, len(allMatch) == 2)
//	assert.Equal(t, s1, allMatch[0])
//	assert.Equal(t, s2, allMatch[1])
//
//	s = `version: 1.1
//stages:
//- stage:
//  - custom-script:
//      commands:
//      - echo maintainer=lj > ${METAFILE}
//  - js:
//- stage:
//  - java:
//      params:
//        invalidref: ${js:OUTPUTS:image} ${js:OUTPUTSS:image}
//        ref: ${js:OUTPUT:image}
//        unknownop: ${js:XXX:images}
//        cs: ${custom-script:OUTPUT:maintainer}
//`
//	allMatch = re.FindAllString(s, -1)
//	spew.Dump(allMatch)
//}
//
//func TestReV2(t *testing.T) {
//	re := expression.Re
//
//	s := `${{ java }}`
//	allMatch := re.FindStringSubmatch(s)
//	assert.True(t, len(allMatch) == 1)
//	assert.Equal(t, s, allMatch[0])
//
//	s = `${{ HOME1:-xxx }}`
//	allMatch = re.FindStringSubmatch(s)
//	assert.True(t, len(allMatch) == 1)
//	assert.Equal(t, s, allMatch[0])
//
//	s = `${{ java:OUTPUT:image }}`
//	allMatch = re.FindStringSubmatch(s)
//	assert.True(t, len(allMatch) == 1)
//	assert.Equal(t, s, allMatch[0])
//
//	s1 := `${{ java }}`
//	s2 := `${{ java:OUTPUT:image }}`
//	s = strutil.Concat(s1, " ", s2)
//	allMatch = re.FindStringSubmatch(s)
//	assert.True(t, len(allMatch) == 2)
//	assert.Equal(t, s1, allMatch[0])
//	assert.Equal(t, s2, allMatch[1])
//
//	s = `version: 1.1
//stages:
//- stage:
//  - custom-script:
//      commands:
//      - echo maintainer=lj > ${{METAFILE}}
//  - js:
//- stage:
//  - java:
//      params:
//        invalidref: ${{js:OUTPUTS:image}} ${{js:OUTPUTSS:image}}
//        ref: ${{js:OUTPUT:image}}
//        ${{}}
//        unknownop: ${{js:XXX:images}}
//        cs: ${{custom-script:OUTPUT:maintainer}}
//`
//
//	strutil.ReplaceAllStringSubmatchFunc(re, s, func(strings []string) string {
//		for _, v := range strings {
//			fmt.Println(v)
//		}
//		return ""
//	})
//	allMatch = re.FindAllString(s, -1)
//	spew.Dump(allMatch)
//}
