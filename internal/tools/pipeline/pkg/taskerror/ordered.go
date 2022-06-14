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

package taskerror

type OrderedResponses []*PipelineTaskErrResponse

func (o OrderedResponses) Len() int           { return len(o) }
func (o OrderedResponses) Less(i, j int) bool { return o[i].Ctx.EndTime.Before(o[j].Ctx.EndTime) }
func (o OrderedResponses) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
