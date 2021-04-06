package tool

func IsKeyInArray(array []string, key string) bool {
	for _, v := range array {
		if v == key {
			return true
		}
	}
	return false
}
