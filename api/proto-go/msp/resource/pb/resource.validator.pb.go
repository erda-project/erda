// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: resource.proto

package pb

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	_ "github.com/mwitkow/go-proto-validators"
	github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
	_ "google.golang.org/genproto/googleapis/api/annotations"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *CreateResourceRequest) Validate() error {
	if this.Engine == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Engine", fmt.Errorf(`value '%v' must not be an empty string`, this.Engine))
	}
	if this.Uuid == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Uuid", fmt.Errorf(`value '%v' must not be an empty string`, this.Uuid))
	}
	if this.Az == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Az", fmt.Errorf(`value '%v' must not be an empty string`, this.Az))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *CreateResourceResponse) Validate() error {
	if this.Data != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Data); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Data", err)
		}
	}
	return nil
}
func (this *ResourceCreateResult) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *DeleteResourceRequest) Validate() error {
	if this.Engine == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Engine", fmt.Errorf(`value '%v' must not be an empty string`, this.Engine))
	}
	if this.Id == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Id", fmt.Errorf(`value '%v' must not be an empty string`, this.Id))
	}
	return nil
}
func (this *DeleteResourceResponse) Validate() error {
	return nil
}
func (this *GetMonitorRuntimeRequest) Validate() error {
	return nil
}
func (this *GetMonitorRuntimeResponse) Validate() error {
	if this.Data != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Data); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Data", err)
		}
	}
	return nil
}
func (this *MonitorRuntime) Validate() error {
	return nil
}
func (this *GetMonitorInstanceRequest) Validate() error {
	if this.TerminusKey == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("TerminusKey", fmt.Errorf(`value '%v' must not be an empty string`, this.TerminusKey))
	}
	return nil
}
func (this *GetMonitorInstanceResponse) Validate() error {
	if this.Data != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Data); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Data", err)
		}
	}
	return nil
}
func (this *MonitorInstance) Validate() error {
	return nil
}