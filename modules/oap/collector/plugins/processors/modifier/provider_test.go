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

package modifier

// func Test_provider_modify(t *testing.T) {
// 	type fields struct {
// 		Cfg *config
// 	}
// 	type args struct {
// 		tags map[string]string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   map[string]string
// 	}{
// 		{
// 			name: "test trim_left",
// 			fields: fields{Cfg: &config{Rules: []ModifierCfg{
// 				{
// 					Action: TrimLeft,
// 					Key:    "kubernetes_pod_",
// 				},
// 				{
// 					Action: TrimLeft,
// 					Key:    "annotations_msp_erda_cloud_",
// 				},
// 				{
// 					Action: TrimLeft,
// 					Key:    "xxx_annotations",
// 				},
// 			}}},
// 			args: args{tags: map[string]string{
// 				metadata.PrefixPodAnnotations + common.NormalizeKey("msp.erda.cloud/application_name"): "test",
// 			}},
// 			want: map[string]string{
// 				"application_name": "test",
// 			},
// 		},
// 		{
// 			name: "test add",
// 			fields: fields{Cfg: &config{Rules: []ModifierCfg{
// 				{
// 					Action: Add,
// 					Key:    "aaa",
// 					Value:  "bbb",
// 				},
// 				{
// 					Action: Add,
// 					Key:    "xxx",
// 					Value:  "yyyyyy",
// 				},
// 			}}},
// 			args: args{tags: map[string]string{
// 				"xxx": "yyy",
// 			}},
// 			want: map[string]string{
// 				"aaa": "bbb",
// 				"xxx": "yyy",
// 			},
// 		},
// 		{
// 			name: "test set",
// 			fields: fields{Cfg: &config{Rules: []ModifierCfg{
// 				{
// 					Action: Set,
// 					Key:    "aaa",
// 					Value:  "bbb",
// 				},
// 				{
// 					Action: Set,
// 					Key:    "xxx",
// 					Value:  "yyyyyy",
// 				},
// 			}}},
// 			args: args{tags: map[string]string{
// 				"xxx": "yyy",
// 			}},
// 			want: map[string]string{
// 				"aaa": "bbb",
// 				"xxx": "yyyyyy",
// 			},
// 		},
// 		{
// 			name: "test rename",
// 			fields: fields{Cfg: &config{Rules: []ModifierCfg{
// 				{
// 					Action: Rename,
// 					Key:    "aaa",
// 					Value:  "bbb",
// 				},
// 				{
// 					Action: Rename,
// 					Key:    "xxx",
// 					Value:  "xxxxxx",
// 				},
// 			}}},
// 			args: args{tags: map[string]string{
// 				"xxx": "yyy",
// 			}},
// 			want: map[string]string{
// 				"xxxxxx": "yyy",
// 			},
// 		},
// 		{
// 			name: "test drop",
// 			fields: fields{Cfg: &config{Rules: []ModifierCfg{
// 				{
// 					Action: Drop,
// 					Key:    "aaa",
// 					Value:  "bbb",
// 				},
// 			}}},
// 			args: args{tags: map[string]string{
// 				"aaa": "bbb",
// 			}},
// 			want: map[string]string{},
// 		},
// 		{
// 			name: "test copy",
// 			fields: fields{Cfg: &config{Rules: []ModifierCfg{
// 				{
// 					Action: Copy,
// 					Key:    "aaa",
// 					Value:  "aaa2",
// 				},
// 			}}},
// 			args: args{tags: map[string]string{
// 				"aaa": "bbb",
// 			}},
// 			want: map[string]string{
// 				"aaa":  "bbb",
// 				"aaa2": "bbb",
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			p := &provider{
// 				Cfg: tt.fields.Cfg,
// 			}
// 			if got := p.modifyTags(tt.args.tags); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("modifyTags() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
