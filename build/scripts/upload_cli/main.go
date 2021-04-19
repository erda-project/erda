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

package main

import (
	"fmt"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func main() {
	keyID := os.Args[1]
	keySecret := os.Args[2]
	objName := os.Args[3]
	file := os.Args[4]
	// 创建OSSClient实例。
	client, err := oss.New("oss-cn-hangzhou.aliyuncs.com", keyID, keySecret)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// 获取存储空间。
	bucket, err := client.Bucket("terminus-dice")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// 上传本地文件。
	err = bucket.PutObjectFromFile(objName, file)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
}
