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

package xmind

import (
	"encoding/xml"
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/mholt/archiver"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/erda-project/erda/pkg/strutil"
)

func TestParseXML(t *testing.T) {
	f, err := os.Open("./examples/xmind/content.xml.ok")
	assert.NoError(t, err)
	m, err := ParseXML(f)
	assert.NoError(t, err)
	spew.Dump(m)
}

func TestExportXMind(t *testing.T) {
	content := XMLContent{
		Sheet: XMLSheet{
			Topic: &XMLTopic{
				Title: "dice-backup",
				Children: &XMLTopicChildren{
					TypedTopics: &XMLTypedTopics{
						Type: TopicsTypeAttached,
						Topics: []*XMLTopic{
							{
								Title: "目录1",
								Children: &XMLTopicChildren{
									TypedTopics: &XMLTypedTopics{
										Type: TopicsTypeAttached,
										Topics: []*XMLTopic{
											{
												Title: "tc:P3__新增应用",
												Children: &XMLTopicChildren{
													TypedTopics: &XMLTypedTopics{
														Type: TopicsTypeAttached,
														Topics: []*XMLTopic{
															{
																Title: "p:已经加入项目",
																Children: &XMLTopicChildren{
																	TypedTopics: &XMLTypedTopics{
																		Type: TopicsTypeAttached,
																		Topics: []*XMLTopic{
																			{
																				Title: "步骤1",
																				Children: &XMLTopicChildren{
																					TypedTopics: &XMLTypedTopics{
																						Type: TopicsTypeAttached,
																						Topics: []*XMLTopic{
																							{
																								Title: "结果1",
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	b, _ := xml.MarshalIndent(&content, "", "  ")
	fmt.Println(string(b))

	f, err := os.OpenFile("/tmp/exported.xmind", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0755)
	assert.NoError(t, err)

	err = Export(f, content, "sfb")
	assert.NoError(t, err)
}

func TestZipXmind(t *testing.T) {
	os.Chdir("./examples/xmind")
	files := []string{"META-INF/", "content.xml", "meta.xml"}
	err := archiver.Zip.Make("a.xmind", files)
	assert.NoError(t, err)
}

// TestAddAttachedChildTopic
func TestAddAttachedChildTopic(t *testing.T) {
	rootTopic := &XMLTopic{
		Title: "测试用例",
	}

	topic1 := rootTopic
	for _, dir := range strutil.Split("/d1", "/", true) {
		topic1 = topic1.AddAttachedChildTopic(dir, true)
	}

	topic2 := rootTopic
	for _, dir := range strutil.Split("/d1/d4", "/", true) {
		topic2 = topic2.AddAttachedChildTopic(dir, true)
	}

	topic3 := topic2
	for _, dir := range strutil.Split("/d5/d6", "/", true) {
		topic3 = topic3.AddAttachedChildTopic(dir, true)
	}

	b, err := yaml.Marshal(rootTopic)
	assert.NoError(t, err)
	fmt.Println(string(b))
}
