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

package marathon

import (
	"fmt"

	"github.com/erda-project/erda/pkg/strutil"
)

// Constraints Used to build & generate marathon constraints label
type Constraints struct {
	//          key: tagname
	likeRules   map[string][]*rule
	unlikeRules map[string][]*rule
}

// NewConstraints create Constraints
func NewConstraints() Constraints {
	return Constraints{likeRules: make(map[string][]*rule), unlikeRules: make(map[string][]*rule)}
}

// NewLikeRule add like rule
func (c *Constraints) NewLikeRule(tag string) *rule {
	r := rule{}
	c.likeRules[tag] = append(c.likeRules[tag], &r)
	return &r
}

// NewUnlikeRule add unlike rule
func (c *Constraints) NewUnlikeRule(tag string) *rule {
	r := rule{}
	c.unlikeRules[tag] = append(c.unlikeRules[tag], &r)
	return &r
}

// Generate Generate Marathon label
func (c *Constraints) Generate() []Constraint {
	r := []Constraint{}
	for tag, likes := range c.likeRules {
		for _, like := range likes {
			r = append(r, Constraint{tag, "LIKE", like.generate()})
		}
	}

	for tag, unlikes := range c.unlikeRules {
		for _, unlike := range unlikes {
			r = append(r, Constraint{tag, "UNLIKE", unlike.generate()})
		}
	}
	return r
}

type rule struct {
	elems []*andrule
}

type andrule struct {
	elems []string
}

// Semantically equal to'|' (or)
func (r *rule) OR(or ...*andrule) *rule {
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

// AND Semantically equal to'&' (and)
func AND(ss ...string) *andrule {
	return &andrule{elems: ss}
}
