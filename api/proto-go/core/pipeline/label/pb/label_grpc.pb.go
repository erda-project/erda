// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// Source: label.proto

package pb

import (
	context "context"

	transport "github.com/erda-project/erda-infra/pkg/transport"
	grpc1 "github.com/erda-project/erda-infra/pkg/transport/grpc"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion5

// LabelServiceClient is the client API for LabelService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LabelServiceClient interface {
	PipelineLabelBatchInsert(ctx context.Context, in *PipelineLabelBatchInsertRequest, opts ...grpc.CallOption) (*PipelineLabelBatchInsertResponse, error)
	PipelineLabelList(ctx context.Context, in *PipelineLabelListRequest, opts ...grpc.CallOption) (*PipelineLabelListResponse, error)
}

type labelServiceClient struct {
	cc grpc1.ClientConnInterface
}

func NewLabelServiceClient(cc grpc1.ClientConnInterface) LabelServiceClient {
	return &labelServiceClient{cc}
}

func (c *labelServiceClient) PipelineLabelBatchInsert(ctx context.Context, in *PipelineLabelBatchInsertRequest, opts ...grpc.CallOption) (*PipelineLabelBatchInsertResponse, error) {
	out := new(PipelineLabelBatchInsertResponse)
	err := c.cc.Invoke(ctx, "/erda.core.pipeline.label.LabelService/PipelineLabelBatchInsert", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *labelServiceClient) PipelineLabelList(ctx context.Context, in *PipelineLabelListRequest, opts ...grpc.CallOption) (*PipelineLabelListResponse, error) {
	out := new(PipelineLabelListResponse)
	err := c.cc.Invoke(ctx, "/erda.core.pipeline.label.LabelService/PipelineLabelList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LabelServiceServer is the server API for LabelService service.
// All implementations should embed UnimplementedLabelServiceServer
// for forward compatibility
type LabelServiceServer interface {
	PipelineLabelBatchInsert(context.Context, *PipelineLabelBatchInsertRequest) (*PipelineLabelBatchInsertResponse, error)
	PipelineLabelList(context.Context, *PipelineLabelListRequest) (*PipelineLabelListResponse, error)
}

// UnimplementedLabelServiceServer should be embedded to have forward compatible implementations.
type UnimplementedLabelServiceServer struct {
}

func (*UnimplementedLabelServiceServer) PipelineLabelBatchInsert(context.Context, *PipelineLabelBatchInsertRequest) (*PipelineLabelBatchInsertResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PipelineLabelBatchInsert not implemented")
}
func (*UnimplementedLabelServiceServer) PipelineLabelList(context.Context, *PipelineLabelListRequest) (*PipelineLabelListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PipelineLabelList not implemented")
}

func RegisterLabelServiceServer(s grpc1.ServiceRegistrar, srv LabelServiceServer, opts ...grpc1.HandleOption) {
	s.RegisterService(_get_LabelService_serviceDesc(srv, opts...), srv)
}

var _LabelService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "erda.core.pipeline.label.LabelService",
	HandlerType: (*LabelServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "label.proto",
}

func _get_LabelService_serviceDesc(srv LabelServiceServer, opts ...grpc1.HandleOption) *grpc.ServiceDesc {
	h := grpc1.DefaultHandleOptions()
	for _, op := range opts {
		op(h)
	}

	_LabelService_PipelineLabelBatchInsert_Handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.PipelineLabelBatchInsert(ctx, req.(*PipelineLabelBatchInsertRequest))
	}
	var _LabelService_PipelineLabelBatchInsert_info transport.ServiceInfo
	if h.Interceptor != nil {
		_LabelService_PipelineLabelBatchInsert_info = transport.NewServiceInfo("erda.core.pipeline.label.LabelService", "PipelineLabelBatchInsert", srv)
		_LabelService_PipelineLabelBatchInsert_Handler = h.Interceptor(_LabelService_PipelineLabelBatchInsert_Handler)
	}

	_LabelService_PipelineLabelList_Handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.PipelineLabelList(ctx, req.(*PipelineLabelListRequest))
	}
	var _LabelService_PipelineLabelList_info transport.ServiceInfo
	if h.Interceptor != nil {
		_LabelService_PipelineLabelList_info = transport.NewServiceInfo("erda.core.pipeline.label.LabelService", "PipelineLabelList", srv)
		_LabelService_PipelineLabelList_Handler = h.Interceptor(_LabelService_PipelineLabelList_Handler)
	}

	var serviceDesc = _LabelService_serviceDesc
	serviceDesc.Methods = []grpc.MethodDesc{
		{
			MethodName: "PipelineLabelBatchInsert",
			Handler: func(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(PipelineLabelBatchInsertRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil && h.Interceptor == nil {
					return srv.(LabelServiceServer).PipelineLabelBatchInsert(ctx, in)
				}
				if h.Interceptor != nil {
					ctx = context.WithValue(ctx, transport.ServiceInfoContextKey, _LabelService_PipelineLabelBatchInsert_info)
				}
				if interceptor == nil {
					return _LabelService_PipelineLabelBatchInsert_Handler(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/erda.core.pipeline.label.LabelService/PipelineLabelBatchInsert",
				}
				return interceptor(ctx, in, info, _LabelService_PipelineLabelBatchInsert_Handler)
			},
		},
		{
			MethodName: "PipelineLabelList",
			Handler: func(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(PipelineLabelListRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil && h.Interceptor == nil {
					return srv.(LabelServiceServer).PipelineLabelList(ctx, in)
				}
				if h.Interceptor != nil {
					ctx = context.WithValue(ctx, transport.ServiceInfoContextKey, _LabelService_PipelineLabelList_info)
				}
				if interceptor == nil {
					return _LabelService_PipelineLabelList_Handler(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/erda.core.pipeline.label.LabelService/PipelineLabelList",
				}
				return interceptor(ctx, in, info, _LabelService_PipelineLabelList_Handler)
			},
		},
	}
	return &serviceDesc
}