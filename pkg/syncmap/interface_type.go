package syncmap

import "fmt"

func MarkInterfaceType(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v := range x {
			ks := fmt.Sprintf("%v", k)
			m[ks] = MarkInterfaceType(v)
		}
		return m
	case []interface{}:
		for i, v := range x {
			x[i] = MarkInterfaceType(v)
		}
	}
	return i
}
