package license

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestParseLicense(t *testing.T) {
	bytes, err := ioutil.ReadFile("license.json")
	if err != nil {
		panic(err)
	}
	licenseKey, err := AesEncrypt(string(bytes), aesKey)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("license_key.txt", []byte(licenseKey), os.ModePerm)
	println(licenseKey)
	license, err := ParseLicense(licenseKey)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", license.Data)
}
