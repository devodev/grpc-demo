// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.21.0-devel
// 	protoc        v3.11.4
// source: hub.proto

package internal

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Client struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name           string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	ConnectionTime string `protobuf:"bytes,2,opt,name=connectionTime,proto3" json:"connectionTime,omitempty"`
	Uptime         string `protobuf:"bytes,3,opt,name=uptime,proto3" json:"uptime,omitempty"`
}

func (x *Client) Reset() {
	*x = Client{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hub_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Client) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Client) ProtoMessage() {}

func (x *Client) ProtoReflect() protoreflect.Message {
	mi := &file_hub_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Client.ProtoReflect.Descriptor instead.
func (*Client) Descriptor() ([]byte, []int) {
	return file_hub_proto_rawDescGZIP(), []int{0}
}

func (x *Client) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Client) GetConnectionTime() string {
	if x != nil {
		return x.ConnectionTime
	}
	return ""
}

func (x *Client) GetUptime() string {
	if x != nil {
		return x.Uptime
	}
	return ""
}

type HubListClientsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HubListClientsRequest) Reset() {
	*x = HubListClientsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hub_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HubListClientsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HubListClientsRequest) ProtoMessage() {}

func (x *HubListClientsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_hub_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HubListClientsRequest.ProtoReflect.Descriptor instead.
func (*HubListClientsRequest) Descriptor() ([]byte, []int) {
	return file_hub_proto_rawDescGZIP(), []int{1}
}

type HubListClientsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count   int64     `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	Clients []*Client `protobuf:"bytes,2,rep,name=clients,proto3" json:"clients,omitempty"`
}

func (x *HubListClientsResponse) Reset() {
	*x = HubListClientsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hub_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HubListClientsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HubListClientsResponse) ProtoMessage() {}

func (x *HubListClientsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_hub_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HubListClientsResponse.ProtoReflect.Descriptor instead.
func (*HubListClientsResponse) Descriptor() ([]byte, []int) {
	return file_hub_proto_rawDescGZIP(), []int{2}
}

func (x *HubListClientsResponse) GetCount() int64 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *HubListClientsResponse) GetClients() []*Client {
	if x != nil {
		return x.Clients
	}
	return nil
}

type HubActivityFeedRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HubActivityFeedRequest) Reset() {
	*x = HubActivityFeedRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hub_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HubActivityFeedRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HubActivityFeedRequest) ProtoMessage() {}

func (x *HubActivityFeedRequest) ProtoReflect() protoreflect.Message {
	mi := &file_hub_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HubActivityFeedRequest.ProtoReflect.Descriptor instead.
func (*HubActivityFeedRequest) Descriptor() ([]byte, []int) {
	return file_hub_proto_rawDescGZIP(), []int{3}
}

type ActivityEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ActivityEvent) Reset() {
	*x = ActivityEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hub_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ActivityEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ActivityEvent) ProtoMessage() {}

func (x *ActivityEvent) ProtoReflect() protoreflect.Message {
	mi := &file_hub_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ActivityEvent.ProtoReflect.Descriptor instead.
func (*ActivityEvent) Descriptor() ([]byte, []int) {
	return file_hub_proto_rawDescGZIP(), []int{4}
}

func (x *ActivityEvent) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_hub_proto protoreflect.FileDescriptor

var file_hub_proto_rawDesc = []byte{
	0x0a, 0x09, 0x68, 0x75, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x22, 0x5c, 0x0a, 0x06, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x26, 0x0a, 0x0e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x75,
	0x70, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x70, 0x74,
	0x69, 0x6d, 0x65, 0x22, 0x17, 0x0a, 0x15, 0x48, 0x75, 0x62, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x5a, 0x0a, 0x16,
	0x48, 0x75, 0x62, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x2a, 0x0a, 0x07,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52,
	0x07, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x18, 0x0a, 0x16, 0x48, 0x75, 0x62, 0x41,
	0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x46, 0x65, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x22, 0x29, 0x0a, 0x0d, 0x41, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0xaa, 0x01,
	0x0a, 0x03, 0x48, 0x75, 0x62, 0x12, 0x50, 0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x73, 0x12, 0x1f, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e,
	0x48, 0x75, 0x62, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2e, 0x48, 0x75, 0x62, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x51, 0x0a, 0x12, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x41, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x46, 0x65, 0x65, 0x64, 0x12, 0x20, 0x2e,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x48, 0x75, 0x62, 0x41, 0x63, 0x74, 0x69,
	0x76, 0x69, 0x74, 0x79, 0x46, 0x65, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x17, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x76,
	0x69, 0x74, 0x79, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x30, 0x01, 0x42, 0x33, 0x5a, 0x31, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x65, 0x76, 0x6f, 0x64, 0x65, 0x76,
	0x2f, 0x67, 0x72, 0x70, 0x63, 0x2d, 0x64, 0x65, 0x6d, 0x6f, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x62, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_hub_proto_rawDescOnce sync.Once
	file_hub_proto_rawDescData = file_hub_proto_rawDesc
)

func file_hub_proto_rawDescGZIP() []byte {
	file_hub_proto_rawDescOnce.Do(func() {
		file_hub_proto_rawDescData = protoimpl.X.CompressGZIP(file_hub_proto_rawDescData)
	})
	return file_hub_proto_rawDescData
}

var file_hub_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_hub_proto_goTypes = []interface{}{
	(*Client)(nil),                 // 0: internal.Client
	(*HubListClientsRequest)(nil),  // 1: internal.HubListClientsRequest
	(*HubListClientsResponse)(nil), // 2: internal.HubListClientsResponse
	(*HubActivityFeedRequest)(nil), // 3: internal.HubActivityFeedRequest
	(*ActivityEvent)(nil),          // 4: internal.ActivityEvent
}
var file_hub_proto_depIdxs = []int32{
	0, // 0: internal.HubListClientsResponse.clients:type_name -> internal.Client
	1, // 1: internal.Hub.ListClients:input_type -> internal.HubListClientsRequest
	3, // 2: internal.Hub.StreamActivityFeed:input_type -> internal.HubActivityFeedRequest
	2, // 3: internal.Hub.ListClients:output_type -> internal.HubListClientsResponse
	4, // 4: internal.Hub.StreamActivityFeed:output_type -> internal.ActivityEvent
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_hub_proto_init() }
func file_hub_proto_init() {
	if File_hub_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_hub_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Client); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_hub_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HubListClientsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_hub_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HubListClientsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_hub_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HubActivityFeedRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_hub_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ActivityEvent); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_hub_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_hub_proto_goTypes,
		DependencyIndexes: file_hub_proto_depIdxs,
		MessageInfos:      file_hub_proto_msgTypes,
	}.Build()
	File_hub_proto = out.File
	file_hub_proto_rawDesc = nil
	file_hub_proto_goTypes = nil
	file_hub_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// HubClient is the client API for Hub service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type HubClient interface {
	ListClients(ctx context.Context, in *HubListClientsRequest, opts ...grpc.CallOption) (*HubListClientsResponse, error)
	StreamActivityFeed(ctx context.Context, in *HubActivityFeedRequest, opts ...grpc.CallOption) (Hub_StreamActivityFeedClient, error)
}

type hubClient struct {
	cc grpc.ClientConnInterface
}

func NewHubClient(cc grpc.ClientConnInterface) HubClient {
	return &hubClient{cc}
}

func (c *hubClient) ListClients(ctx context.Context, in *HubListClientsRequest, opts ...grpc.CallOption) (*HubListClientsResponse, error) {
	out := new(HubListClientsResponse)
	err := c.cc.Invoke(ctx, "/internal.Hub/ListClients", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) StreamActivityFeed(ctx context.Context, in *HubActivityFeedRequest, opts ...grpc.CallOption) (Hub_StreamActivityFeedClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Hub_serviceDesc.Streams[0], "/internal.Hub/StreamActivityFeed", opts...)
	if err != nil {
		return nil, err
	}
	x := &hubStreamActivityFeedClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Hub_StreamActivityFeedClient interface {
	Recv() (*ActivityEvent, error)
	grpc.ClientStream
}

type hubStreamActivityFeedClient struct {
	grpc.ClientStream
}

func (x *hubStreamActivityFeedClient) Recv() (*ActivityEvent, error) {
	m := new(ActivityEvent)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// HubServer is the server API for Hub service.
type HubServer interface {
	ListClients(context.Context, *HubListClientsRequest) (*HubListClientsResponse, error)
	StreamActivityFeed(*HubActivityFeedRequest, Hub_StreamActivityFeedServer) error
}

// UnimplementedHubServer can be embedded to have forward compatible implementations.
type UnimplementedHubServer struct {
}

func (*UnimplementedHubServer) ListClients(context.Context, *HubListClientsRequest) (*HubListClientsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListClients not implemented")
}
func (*UnimplementedHubServer) StreamActivityFeed(*HubActivityFeedRequest, Hub_StreamActivityFeedServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamActivityFeed not implemented")
}

func RegisterHubServer(s *grpc.Server, srv HubServer) {
	s.RegisterService(&_Hub_serviceDesc, srv)
}

func _Hub_ListClients_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HubListClientsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).ListClients(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/internal.Hub/ListClients",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).ListClients(ctx, req.(*HubListClientsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_StreamActivityFeed_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(HubActivityFeedRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HubServer).StreamActivityFeed(m, &hubStreamActivityFeedServer{stream})
}

type Hub_StreamActivityFeedServer interface {
	Send(*ActivityEvent) error
	grpc.ServerStream
}

type hubStreamActivityFeedServer struct {
	grpc.ServerStream
}

func (x *hubStreamActivityFeedServer) Send(m *ActivityEvent) error {
	return x.ServerStream.SendMsg(m)
}

var _Hub_serviceDesc = grpc.ServiceDesc{
	ServiceName: "internal.Hub",
	HandlerType: (*HubServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListClients",
			Handler:    _Hub_ListClients_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamActivityFeed",
			Handler:       _Hub_StreamActivityFeed_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "hub.proto",
}
