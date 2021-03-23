package kmstypes

import (
	"regexp"
)

const kindNameFormat = `^[A-Z0-9_]+$`

var formatter = regexp.MustCompile(kindNameFormat)

type PluginKind string

func (s PluginKind) String() string {
	return string(s)
}

func (s PluginKind) Validate() bool {
	return formatter.MatchString(string(s))
}

type StoreKind string

func (s StoreKind) String() string {
	return string(s)
}

func (s StoreKind) Validate() bool {
	return formatter.MatchString(string(s))
}
