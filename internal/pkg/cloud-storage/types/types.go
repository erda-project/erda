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

package types

import (
	"time"
)

type PutObjectResult struct {
	// Content-Md5 for the uploaded object.
	ContentMD5 string

	// Entity tag for the uploaded object.
	ETag string

	// The 64-bit CRC value of the object.
	// This value is calculated based on the ECMA-182 standard.
	HashCRC64 string

	// Version of the object.
	VersionId string

	Bucket string

	ObjectName string
	Region     string
}

type FileObject struct {
	Key          string
	Type         *string    `xml:"Type"`
	Size         int64      `xml:"Size"`
	LastModified *time.Time `xml:"LastModified"`
}
