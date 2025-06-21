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

package i18n

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/i18n/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
)

// Config AI proxy internationalization configuration table
type Config struct {
	common.BaseModel

	Category  string `gorm:"column:category;type:varchar(50);not null;comment:configuration category"`
	ItemKey   string `gorm:"column:item_key;type:varchar(100);not null;comment:item identifier"`
	FieldName string `gorm:"column:field_name;type:varchar(50);not null;comment:field name"`
	Locale    string `gorm:"column:locale;type:varchar(10);not null;comment:language identifier"`
	Value     string `gorm:"column:value;type:text;not null;comment:configuration value"`
}

// TableName table name
func (Config) TableName() string {
	return "ai_proxy_i18n"
}

// ToProtobuf convert to protobuf format
func (c *Config) ToProtobuf() *pb.I18NConfig {
	return &pb.I18NConfig{
		Id:        c.ID.String,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
		DeletedAt: timestamppb.New(c.DeletedAt.Time),
		Category:  c.Category,
		ItemKey:   c.ItemKey,
		FieldName: c.FieldName,
		Locale:    c.Locale,
		Value:     c.Value,
	}
}

// ConfigCategory configuration category constants
type ConfigCategory string

const (
	CategoryPublisher ConfigCategory = "publisher" // publisher configuration
	CategoryAbility   ConfigCategory = "ability"   // ability configuration
)

// FieldName field name constants
type FieldName string

const (
	FieldNameValue FieldName = "name" // name
	FieldLogo      FieldName = "logo" // logo URL
)

// Locale language constants
type Locale string

const (
	LocaleUniversal Locale = "*"  // universal configuration (e.g. logo)
	LocaleEnglish   Locale = "en" // English
	LocaleChinese   Locale = "zh" // Chinese
	LocaleDefault   Locale = LocaleChinese
)

// PublisherKey publisher identifier constants
type PublisherKey string
