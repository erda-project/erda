package marathon

import (
	"fmt"

	"github.com/erda-project/erda/pkg/strutil"
)

func (t *constraintTerm) generate() []string {
	r := rule{}
	for _, v := range t.Values {
		r.or(and(v))
	}
	return []string{t.Key, string(t.Op), r.generate()}
}

type rule struct {
	elems []*andrule
}

type andrule struct {
	elems []string
}

// OR 语义上等于 '|' （或）
func (r *rule) or(or ...*andrule) *rule {
	r.elems = append(r.elems, or...)
	return r
}

func (r *andrule) generate() string {
	if len(r.elems) == 0 {
		return ""
	}
	if len(r.elems) == 1 {
		return strutil.Concat(r.elems[0], `\b.*`)
	}
	rules := []string{}
	for _, rule := range r.elems {
		rules = append(rules, strutil.Concat(rule, `\b.*`))
	}
	return strutil.Concat("(", strutil.Join(rules, "|"), ")", fmt.Sprintf("{%d}", len(r.elems)))
}

func (r *rule) generate() string {
	andrules := []string{}
	for _, andrule := range r.elems {
		andrules = append(andrules, andrule.generate())
	}
	return strutil.Concat(`.*\b(`, strutil.Join(andrules, "|", true), ")")
}

// AND 语义上等于 '&' (且)
func and(ss ...string) *andrule {
	return &andrule{elems: ss}
}
