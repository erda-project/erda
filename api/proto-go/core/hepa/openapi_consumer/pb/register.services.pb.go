// Code generated by protoc-gen-go-register. DO NOT EDIT.
// Sources: openapi_consumer.proto

package pb

import (
	reflect "reflect"

	transport "github.com/erda-project/erda-infra/pkg/transport"
)

// RegisterOpenapiConsumerServiceImp openapi_consumer.proto
func RegisterOpenapiConsumerServiceImp(regester transport.Register, srv OpenapiConsumerServiceServer, opts ...transport.ServiceOption) {
	_ops := transport.DefaultServiceOptions()
	for _, op := range opts {
		op(_ops)
	}
	RegisterOpenapiConsumerServiceHandler(regester, OpenapiConsumerServiceHandler(srv), _ops.HTTP...)
	RegisterOpenapiConsumerServiceServer(regester, srv, _ops.GRPC...)
}

// ServiceNames return all service names
func ServiceNames(svr ...string) []string {
	return append(svr,
		"erda.core.hepa.openapi_consumer.OpenapiConsumerService",
	)
}

var (
	openapiConsumerServiceClientType  = reflect.TypeOf((*OpenapiConsumerServiceClient)(nil)).Elem()
	openapiConsumerServiceServerType  = reflect.TypeOf((*OpenapiConsumerServiceServer)(nil)).Elem()
	openapiConsumerServiceHandlerType = reflect.TypeOf((*OpenapiConsumerServiceHandler)(nil)).Elem()
)

// OpenapiConsumerServiceClientType .
func OpenapiConsumerServiceClientType() reflect.Type { return openapiConsumerServiceClientType }

// OpenapiConsumerServiceServerType .
func OpenapiConsumerServiceServerType() reflect.Type { return openapiConsumerServiceServerType }

// OpenapiConsumerServiceHandlerType .
func OpenapiConsumerServiceHandlerType() reflect.Type { return openapiConsumerServiceHandlerType }

func Types() []reflect.Type {
	return []reflect.Type{
		// client types
		openapiConsumerServiceClientType,
		// server types
		openapiConsumerServiceServerType,
		// handler types
		openapiConsumerServiceHandlerType,
	}
}