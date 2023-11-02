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

package message

type Group struct {
	AllMessages Messages // use this directly if you are not care about the details

	// pay attention to the order
	SystemMessage           *Message
	SessionTopicMessage     *Message
	PromptTemplateMessages  Messages
	SessionPreviousMessages Messages // from prompt template if provided in the http header
	RequestedMessages       Messages // normally from body prompt field, user input
}
