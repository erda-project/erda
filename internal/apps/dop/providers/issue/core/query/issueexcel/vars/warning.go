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

package vars

func (data *DataForFulfill) AppendImportWarning(lineNum int, warning string) {
	r := Record{
		LineNum: lineNum,
		Msg:     warning,
	}
	data.ImportOnly.Warnings = append(data.ImportOnly.Warnings, r)
}

func (data *DataForFulfill) AppendImportError(lineNum int, err string) {
	r := Record{
		LineNum: lineNum,
		Msg:     err,
	}
	data.ImportOnly.Errs = append(data.ImportOnly.Errs, r)
}
