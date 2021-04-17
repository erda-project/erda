package convert

import "strconv"

// Converters .
var Converters = map[string]func(text string) (interface{}, error){
	"string": func(text string) (interface{}, error) {
		return text, nil
	},
	"number": func(text string) (interface{}, error) {
		return strconv.ParseFloat(text, 64)
	},
	"bool": func(text string) (interface{}, error) {
		return strconv.ParseBool(text)
	},
	"timestamp": func(text string) (interface{}, error) {
		return strconv.ParseInt(text, 10, 64)
	},
	"int": func(text string) (interface{}, error) {
		return strconv.ParseInt(text, 10, 64)
	},
	"float": func(text string) (interface{}, error) {
		return strconv.ParseFloat(text, 64)
	},
}

// Converter .
func Converter(typ string) func(text string) (interface{}, error) {
	c, ok := Converters[typ]
	if !ok {
		return Converters["string"]
	}
	return c
}
