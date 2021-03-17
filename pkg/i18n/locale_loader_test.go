package i18n

import (
	"testing"
)

func TestLocaleLoader(t *testing.T) {
	loader := NewLoader()
	err := loader.LoadFile("test-locale1.json", "test-locale2.json", "test-locale.yml")
	if err != nil {
		panic(err)
	}
	println(loader.Locale(ZH).Get("dice.not_login"))
	println(loader.Locale(EN).Get("dice.not_login"))
	println(loader.Locale(ZH).Get("dice.multiline"))

}

func TestLocaleLoaderTemplate(t *testing.T) {
	loader := NewLoader()
	err := loader.LoadFile("test-locale1.json", "test-locale2.json", "test-locale.yml")
	if err != nil {
		panic(err)
	}
	template := loader.Locale(ZH).GetTemplate("dice.resource_not_found")
	println(template.RenderByKey(map[string]string{
		"name": "11",
	}))
}
