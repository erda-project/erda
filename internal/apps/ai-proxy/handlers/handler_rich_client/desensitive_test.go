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

package handler_rich_client

import (
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	serviceproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
)

func TestDesensitiveModel(t *testing.T) {
	// Create test metadata with both public and secret data
	publicData, _ := structpb.NewStruct(map[string]interface{}{
		"endpoint": "https://api.example.com",
		"version":  "v1",
	})
	secretData, _ := structpb.NewStruct(map[string]interface{}{
		"secret_key": "super-secret-key",
		"password":   "secret-password",
	})

	// Create test model with sensitive data
	model := &modelpb.Model{
		Id:         "test-model-id",
		CreatedAt:  timestamppb.Now(),
		UpdatedAt:  timestamppb.Now(),
		Name:       "Test Model",
		Desc:       "Test Description",
		Type:       modelpb.ModelType_text_generation,
		ProviderId: "test-provider-id",
		ApiKey:     "sensitive-api-key",
		Publisher:  "Test Publisher",
		Metadata: &metadatapb.Metadata{
			Public: publicData.Fields,
			Secret: secretData.Fields,
		},
	}

	// Test the desensitiveModel function
	desensitiveModel(model)

	// Verify sensitive data is removed
	if model.ApiKey != "" {
		t.Errorf("Expected ApiKey to be empty, got: %s", model.ApiKey)
	}
	if model.Metadata.Secret != nil {
		t.Errorf("Expected Metadata.Secret to be nil, got: %v", model.Metadata.Secret)
	}

	// Verify public data is preserved
	if model.Metadata.Public == nil {
		t.Error("Expected Metadata.Public to be preserved")
	}

	// Verify other fields are preserved
	if model.Name != "Test Model" {
		t.Errorf("Expected Name to be preserved, got: %s", model.Name)
	}
	if model.Publisher != "Test Publisher" {
		t.Errorf("Expected Publisher to be preserved, got: %s", model.Publisher)
	}
}

func TestDesensitiveProvider(t *testing.T) {
	// Create test metadata with both public and secret data
	publicData, _ := structpb.NewStruct(map[string]interface{}{
		"endpoint": "https://api.provider.com",
		"region":   "us-east-1",
	})
	secretData, _ := structpb.NewStruct(map[string]interface{}{
		"api_secret": "provider-secret",
		"token":      "secret-token",
	})

	// Create test provider with sensitive data
	provider := &serviceproviderpb.ServiceProvider{
		Id:        "test-provider-id",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Name:      "Test Provider",
		Desc:      "Test Provider Description",
		Type:      "OpenAI",
		ApiKey:    "sensitive-provider-key",
		Metadata: &metadatapb.Metadata{
			Public: publicData.Fields,
			Secret: secretData.Fields,
		},
	}

	// Test the desensitiveProvider function
	desensitiveProvider(provider)

	// Verify sensitive data is removed
	if provider.ApiKey != "" {
		t.Errorf("Expected ApiKey to be empty, got: %s", provider.ApiKey)
	}
	if provider.Metadata.Secret != nil {
		t.Errorf("Expected Metadata.Secret to be nil, got: %v", provider.Metadata.Secret)
	}

	// Verify public data is preserved
	if provider.Metadata.Public == nil {
		t.Error("Expected Metadata.Public to be preserved")
	}

	// Verify other fields are preserved
	if provider.Name != "Test Provider" {
		t.Errorf("Expected Name to be preserved, got: %s", provider.Name)
	}
	if provider.Type != "OpenAI" {
		t.Errorf("Expected Type to be preserved, got: %s", provider.Type)
	}
}

func TestDesensitiveModel_NilInput(t *testing.T) {
	// Should not panic with nil input
	desensitiveModel(nil)
}

func TestDesensitiveProvider_NilInput(t *testing.T) {
	// Should not panic with nil input
	desensitiveProvider(nil)
}

func TestDesensitiveModel_NilMetadata(t *testing.T) {
	model := &modelpb.Model{
		Id:       "test-model-id",
		Name:     "Test Model",
		ApiKey:   "sensitive-api-key",
		Metadata: nil,
	}

	desensitiveModel(model)

	if model.ApiKey != "" {
		t.Errorf("Expected ApiKey to be empty, got: %s", model.ApiKey)
	}
	if model.Metadata != nil {
		t.Error("Expected Metadata to remain nil")
	}
}

func TestDesensitiveProvider_NilMetadata(t *testing.T) {
	provider := &serviceproviderpb.ServiceProvider{
		Id:       "test-provider-id",
		Name:     "Test Provider",
		ApiKey:   "sensitive-provider-key",
		Metadata: nil,
	}

	desensitiveProvider(provider)

	if provider.ApiKey != "" {
		t.Errorf("Expected ApiKey to be empty, got: %s", provider.ApiKey)
	}
	if provider.Metadata != nil {
		t.Error("Expected Metadata to remain nil")
	}
}

func TestDesensitiveClient(t *testing.T) {
	// Create test metadata with both public and secret data
	publicData, _ := structpb.NewStruct(map[string]interface{}{
		"domain": "example.com",
		"region": "us-west-1",
	})
	secretData, _ := structpb.NewStruct(map[string]interface{}{
		"internal_key": "client-secret",
		"token":        "access-token",
	})

	// Create test client with sensitive data
	client := &clientpb.Client{
		Id:          "test-client-id",
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
		Name:        "Test Client",
		Desc:        "Test Client Description",
		AccessKeyId: "sensitive-access-key",
		SecretKeyId: "sensitive-secret-key",
		Metadata: &metadatapb.Metadata{
			Public: publicData.Fields,
			Secret: secretData.Fields,
		},
	}

	// Test the desensitiveClient function
	result := desensitiveClient(client)

	// Verify function returns the same instance
	if result != client {
		t.Error("Expected function to return the same client instance")
	}

	// Verify sensitive data is removed
	if client.AccessKeyId != "" {
		t.Errorf("Expected AccessKeyId to be empty, got: %s", client.AccessKeyId)
	}
	if client.SecretKeyId != "" {
		t.Errorf("Expected SecretKeyId to be empty, got: %s", client.SecretKeyId)
	}
	if client.Metadata.Secret != nil {
		t.Errorf("Expected Metadata.Secret to be nil, got: %v", client.Metadata.Secret)
	}

	// Verify public data is preserved
	if client.Metadata.Public == nil {
		t.Error("Expected Metadata.Public to be preserved")
	}

	// Verify other fields are preserved
	if client.Name != "Test Client" {
		t.Errorf("Expected Name to be preserved, got: %s", client.Name)
	}
	if client.Desc != "Test Client Description" {
		t.Errorf("Expected Desc to be preserved, got: %s", client.Desc)
	}
}

func TestDesensitiveClient_NilInput(t *testing.T) {
	// Should not panic with nil input
	result := desensitiveClient(nil)
	if result != nil {
		t.Error("Expected nil result for nil input")
	}
}

func TestDesensitiveClient_NilMetadata(t *testing.T) {
	client := &clientpb.Client{
		Id:          "test-client-id",
		Name:        "Test Client",
		AccessKeyId: "sensitive-access-key",
		SecretKeyId: "sensitive-secret-key",
		Metadata:    nil,
	}

	desensitiveClient(client)

	if client.AccessKeyId != "" {
		t.Errorf("Expected AccessKeyId to be empty, got: %s", client.AccessKeyId)
	}
	if client.SecretKeyId != "" {
		t.Errorf("Expected SecretKeyId to be empty, got: %s", client.SecretKeyId)
	}
	if client.Metadata != nil {
		t.Error("Expected Metadata to remain nil")
	}
}

func TestDesensitiveProvider_ApiKeyDeletion(t *testing.T) {
	// Create test metadata with api key in public data
	publicData, _ := structpb.NewStruct(map[string]interface{}{
		"endpoint": "https://api.provider.com",
		"api":      map[string]interface{}{"key": "should-be-deleted"},
		"region":   "us-east-1",
	})
	secretData, _ := structpb.NewStruct(map[string]interface{}{
		"api_secret": "provider-secret",
	})

	// Create test provider with api config in public metadata
	provider := &serviceproviderpb.ServiceProvider{
		Id:     "test-provider-id",
		Name:   "Test Provider",
		ApiKey: "sensitive-provider-key",
		Metadata: &metadatapb.Metadata{
			Public: publicData.Fields,
			Secret: secretData.Fields,
		},
	}

	// Test the desensitiveProvider function
	result := desensitiveProvider(provider)

	// Verify function returns the same instance
	if result != provider {
		t.Error("Expected function to return the same provider instance")
	}

	// Verify api key is deleted from public metadata
	if _, exists := provider.Metadata.Public["api"]; exists {
		t.Error("Expected 'api' key to be deleted from public metadata")
	}

	// Verify other public data is preserved
	if provider.Metadata.Public["endpoint"] == nil {
		t.Error("Expected 'endpoint' to be preserved in public metadata")
	}
	if provider.Metadata.Public["region"] == nil {
		t.Error("Expected 'region' to be preserved in public metadata")
	}
}
