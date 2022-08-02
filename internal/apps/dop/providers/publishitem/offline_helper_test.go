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

package publishitem

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/apps/dop/providers/publishitem/db"
)

func Test_isOffLineVersion(t *testing.T) {
	tests := []struct {
		name    string
		want    bool
		version *db.PublishItemVersion
	}{
		{
			name: "offline",
			want: true,
			version: &db.PublishItemVersion{
				Meta: `{"projectName": "OFFLINE"}`,
			},
		},
		{
			name: "online",
			want: false,
			version: &db.PublishItemVersion{
				Meta: `{"projectName": "ONLINE"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOffLineVersion(tt.version); got != tt.want {
				t.Errorf("isOffLineVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAppStoreURL(t *testing.T) {
	bundleID := "1"
	url := getAppStoreURL(bundleID)
	assert.Equal(t, "", url)
}

func TestGenrateTmpImagePath(t *testing.T) {
	id := "1"
	path := GenrateTmpImagePath(id)
	assert.Equal(t, "/tmp/logo-1.jpg", path)
}

func TestGenerateInstallPlist(t *testing.T) {
	plist := GenerateInstallPlist(&IosAppInfo{
		BundleId: "1",
		Version:  "1.0",
		Name:     "ios",
	}, "https://erda.cloud")
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
   <key>items</key>
   <array>
       <dict>
           <key>assets</key>
           <array>
               <dict>
                   <key>kind</key>
                   <string>software-package</string>
                   <key>url</key>
                   <string>https://erda.cloud</string>
               </dict>
           </array>
           <key>metadata</key>
           <dict>
               <key>bundle-identifier</key>
               <string>1</string>
               <key>bundle-version</key>
               <string>1.0</string>
               <key>kind</key>
               <string>software</string>
               <key>subtitle</key>
               <string>ios</string>
               <key>title</key>
               <string>ios</string>
           </dict>
       </dict>
   </array>
</dict>
</plist>`, plist)
}
