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
