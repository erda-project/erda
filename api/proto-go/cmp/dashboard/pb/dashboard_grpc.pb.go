// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// Source: dashboard.proto

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

// ClusterResourceClient is the client API for ClusterResource service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClusterResourceClient interface {
	GetClustersResources(ctx context.Context, in *GetClustersResourcesRequest, opts ...grpc.CallOption) (*GetClusterResourcesResponse, error)
	GetNamespacesResources(ctx context.Context, in *GetNamespacesResourcesRequest, opts ...grpc.CallOption) (*GetNamespacesResourcesResponse, error)
	GetPodsByLabels(ctx context.Context, in *GetPodsByLabelsRequest, opts ...grpc.CallOption) (*GetPodsByLabelsResponse, error)
}

type clusterResourceClient struct {
	cc grpc1.ClientConnInterface
}

func NewClusterResourceClient(cc grpc1.ClientConnInterface) ClusterResourceClient {
	return &clusterResourceClient{cc}
}

func (c *clusterResourceClient) GetClustersResources(ctx context.Context, in *GetClustersResourcesRequest, opts ...grpc.CallOption) (*GetClusterResourcesResponse, error) {
	out := new(GetClusterResourcesResponse)
	err := c.cc.Invoke(ctx, "/erda.cmp.dashboard.resource.ClusterResource/GetClustersResources", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterResourceClient) GetNamespacesResources(ctx context.Context, in *GetNamespacesResourcesRequest, opts ...grpc.CallOption) (*GetNamespacesResourcesResponse, error) {
	out := new(GetNamespacesResourcesResponse)
	err := c.cc.Invoke(ctx, "/erda.cmp.dashboard.resource.ClusterResource/GetNamespacesResources", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterResourceClient) GetPodsByLabels(ctx context.Context, in *GetPodsByLabelsRequest, opts ...grpc.CallOption) (*GetPodsByLabelsResponse, error) {
	out := new(GetPodsByLabelsResponse)
	err := c.cc.Invoke(ctx, "/erda.cmp.dashboard.resource.ClusterResource/GetPodsByLabels", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClusterResourceServer is the server API for ClusterResource service.
// All implementations should embed UnimplementedClusterResourceServer
// for forward compatibility
type ClusterResourceServer interface {
	GetClustersResources(context.Context, *GetClustersResourcesRequest) (*GetClusterResourcesResponse, error)
	GetNamespacesResources(context.Context, *GetNamespacesResourcesRequest) (*GetNamespacesResourcesResponse, error)
	GetPodsByLabels(context.Context, *GetPodsByLabelsRequest) (*GetPodsByLabelsResponse, error)
}

// UnimplementedClusterResourceServer should be embedded to have forward compatible implementations.
type UnimplementedClusterResourceServer struct {
}

func (*UnimplementedClusterResourceServer) GetClustersResources(context.Context, *GetClustersResourcesRequest) (*GetClusterResourcesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetClustersResources not implemented")
}
func (*UnimplementedClusterResourceServer) GetNamespacesResources(context.Context, *GetNamespacesResourcesRequest) (*GetNamespacesResourcesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNamespacesResources not implemented")
}
func (*UnimplementedClusterResourceServer) GetPodsByLabels(context.Context, *GetPodsByLabelsRequest) (*GetPodsByLabelsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPodsByLabels not implemented")
}

func RegisterClusterResourceServer(s grpc1.ServiceRegistrar, srv ClusterResourceServer, opts ...grpc1.HandleOption) {
	s.RegisterService(_get_ClusterResource_serviceDesc(srv, opts...), srv)
}

var _ClusterResource_serviceDesc = grpc.ServiceDesc{
	ServiceName: "erda.cmp.dashboard.resource.ClusterResource",
	HandlerType: (*ClusterResourceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "dashboard.proto",
}

func _get_ClusterResource_serviceDesc(srv ClusterResourceServer, opts ...grpc1.HandleOption) *grpc.ServiceDesc {
	h := grpc1.DefaultHandleOptions()
	for _, op := range opts {
		op(h)
	}

	_ClusterResource_GetClustersResources_Handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.GetClustersResources(ctx, req.(*GetClustersResourcesRequest))
	}
	var _ClusterResource_GetClustersResources_info transport.ServiceInfo
	if h.Interceptor != nil {
		_ClusterResource_GetClustersResources_info = transport.NewServiceInfo("erda.cmp.dashboard.resource.ClusterResource", "GetClustersResources", srv)
		_ClusterResource_GetClustersResources_Handler = h.Interceptor(_ClusterResource_GetClustersResources_Handler)
	}

	_ClusterResource_GetNamespacesResources_Handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.GetNamespacesResources(ctx, req.(*GetNamespacesResourcesRequest))
	}
	var _ClusterResource_GetNamespacesResources_info transport.ServiceInfo
	if h.Interceptor != nil {
		_ClusterResource_GetNamespacesResources_info = transport.NewServiceInfo("erda.cmp.dashboard.resource.ClusterResource", "GetNamespacesResources", srv)
		_ClusterResource_GetNamespacesResources_Handler = h.Interceptor(_ClusterResource_GetNamespacesResources_Handler)
	}

	_ClusterResource_GetPodsByLabels_Handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.GetPodsByLabels(ctx, req.(*GetPodsByLabelsRequest))
	}
	var _ClusterResource_GetPodsByLabels_info transport.ServiceInfo
	if h.Interceptor != nil {
		_ClusterResource_GetPodsByLabels_info = transport.NewServiceInfo("erda.cmp.dashboard.resource.ClusterResource", "GetPodsByLabels", srv)
		_ClusterResource_GetPodsByLabels_Handler = h.Interceptor(_ClusterResource_GetPodsByLabels_Handler)
	}

	var serviceDesc = _ClusterResource_serviceDesc
	serviceDesc.Methods = []grpc.MethodDesc{
		{
			MethodName: "GetClustersResources",
			Handler: func(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(GetClustersResourcesRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil && h.Interceptor == nil {
					return srv.(ClusterResourceServer).GetClustersResources(ctx, in)
				}
				if h.Interceptor != nil {
					ctx = context.WithValue(ctx, transport.ServiceInfoContextKey, _ClusterResource_GetClustersResources_info)
				}
				if interceptor == nil {
					return _ClusterResource_GetClustersResources_Handler(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/erda.cmp.dashboard.resource.ClusterResource/GetClustersResources",
				}
				return interceptor(ctx, in, info, _ClusterResource_GetClustersResources_Handler)
			},
		},
		{
			MethodName: "GetNamespacesResources",
			Handler: func(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(GetNamespacesResourcesRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil && h.Interceptor == nil {
					return srv.(ClusterResourceServer).GetNamespacesResources(ctx, in)
				}
				if h.Interceptor != nil {
					ctx = context.WithValue(ctx, transport.ServiceInfoContextKey, _ClusterResource_GetNamespacesResources_info)
				}
				if interceptor == nil {
					return _ClusterResource_GetNamespacesResources_Handler(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/erda.cmp.dashboard.resource.ClusterResource/GetNamespacesResources",
				}
				return interceptor(ctx, in, info, _ClusterResource_GetNamespacesResources_Handler)
			},
		},
		{
			MethodName: "GetPodsByLabels",
			Handler: func(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(GetPodsByLabelsRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil && h.Interceptor == nil {
					return srv.(ClusterResourceServer).GetPodsByLabels(ctx, in)
				}
				if h.Interceptor != nil {
					ctx = context.WithValue(ctx, transport.ServiceInfoContextKey, _ClusterResource_GetPodsByLabels_info)
				}
				if interceptor == nil {
					return _ClusterResource_GetPodsByLabels_Handler(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/erda.cmp.dashboard.resource.ClusterResource/GetPodsByLabels",
				}
				return interceptor(ctx, in, info, _ClusterResource_GetPodsByLabels_Handler)
			},
		},
	}
	return &serviceDesc
}