// Code generated by protoc-gen-go. DO NOT EDIT.
// source: staticdataservice.proto

package staticdataservice

import (
	context "context"
	fmt "fmt"
	"github.com/ettec/otp-common/model"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ListingId struct {
	ListingId            int32    `protobuf:"varint,1,opt,name=listingId,proto3" json:"listingId,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListingId) Reset()         { *m = ListingId{} }
func (m *ListingId) String() string { return proto.CompactTextString(m) }
func (*ListingId) ProtoMessage()    {}
func (*ListingId) Descriptor() ([]byte, []int) {
	return fileDescriptor_bda44339ea58dbd5, []int{0}
}

func (m *ListingId) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListingId.Unmarshal(m, b)
}
func (m *ListingId) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListingId.Marshal(b, m, deterministic)
}
func (m *ListingId) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListingId.Merge(m, src)
}
func (m *ListingId) XXX_Size() int {
	return xxx_messageInfo_ListingId.Size(m)
}
func (m *ListingId) XXX_DiscardUnknown() {
	xxx_messageInfo_ListingId.DiscardUnknown(m)
}

var xxx_messageInfo_ListingId proto.InternalMessageInfo

func (m *ListingId) GetListingId() int32 {
	if m != nil {
		return m.ListingId
	}
	return 0
}

type ListingIds struct {
	ListingIds           []int32  `protobuf:"varint,1,rep,packed,name=listingIds,proto3" json:"listingIds,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListingIds) Reset()         { *m = ListingIds{} }
func (m *ListingIds) String() string { return proto.CompactTextString(m) }
func (*ListingIds) ProtoMessage()    {}
func (*ListingIds) Descriptor() ([]byte, []int) {
	return fileDescriptor_bda44339ea58dbd5, []int{1}
}

func (m *ListingIds) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListingIds.Unmarshal(m, b)
}
func (m *ListingIds) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListingIds.Marshal(b, m, deterministic)
}
func (m *ListingIds) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListingIds.Merge(m, src)
}
func (m *ListingIds) XXX_Size() int {
	return xxx_messageInfo_ListingIds.Size(m)
}
func (m *ListingIds) XXX_DiscardUnknown() {
	xxx_messageInfo_ListingIds.DiscardUnknown(m)
}

var xxx_messageInfo_ListingIds proto.InternalMessageInfo

func (m *ListingIds) GetListingIds() []int32 {
	if m != nil {
		return m.ListingIds
	}
	return nil
}

type Listings struct {
	Listings             []*model.Listing `protobuf:"bytes,1,rep,name=listings,proto3" json:"listings,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *Listings) Reset()         { *m = Listings{} }
func (m *Listings) String() string { return proto.CompactTextString(m) }
func (*Listings) ProtoMessage()    {}
func (*Listings) Descriptor() ([]byte, []int) {
	return fileDescriptor_bda44339ea58dbd5, []int{2}
}

func (m *Listings) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Listings.Unmarshal(m, b)
}
func (m *Listings) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Listings.Marshal(b, m, deterministic)
}
func (m *Listings) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Listings.Merge(m, src)
}
func (m *Listings) XXX_Size() int {
	return xxx_messageInfo_Listings.Size(m)
}
func (m *Listings) XXX_DiscardUnknown() {
	xxx_messageInfo_Listings.DiscardUnknown(m)
}

var xxx_messageInfo_Listings proto.InternalMessageInfo

func (m *Listings) GetListings() []*model.Listing {
	if m != nil {
		return m.Listings
	}
	return nil
}

type MatchParameters struct {
	SymbolMatch          string   `protobuf:"bytes,1,opt,name=symbolMatch,proto3" json:"symbolMatch,omitempty"`
	NameMatch            string   `protobuf:"bytes,2,opt,name=nameMatch,proto3" json:"nameMatch,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MatchParameters) Reset()         { *m = MatchParameters{} }
func (m *MatchParameters) String() string { return proto.CompactTextString(m) }
func (*MatchParameters) ProtoMessage()    {}
func (*MatchParameters) Descriptor() ([]byte, []int) {
	return fileDescriptor_bda44339ea58dbd5, []int{3}
}

func (m *MatchParameters) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MatchParameters.Unmarshal(m, b)
}
func (m *MatchParameters) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MatchParameters.Marshal(b, m, deterministic)
}
func (m *MatchParameters) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MatchParameters.Merge(m, src)
}
func (m *MatchParameters) XXX_Size() int {
	return xxx_messageInfo_MatchParameters.Size(m)
}
func (m *MatchParameters) XXX_DiscardUnknown() {
	xxx_messageInfo_MatchParameters.DiscardUnknown(m)
}

var xxx_messageInfo_MatchParameters proto.InternalMessageInfo

func (m *MatchParameters) GetSymbolMatch() string {
	if m != nil {
		return m.SymbolMatch
	}
	return ""
}

func (m *MatchParameters) GetNameMatch() string {
	if m != nil {
		return m.NameMatch
	}
	return ""
}

type ExactMatchParameters struct {
	Symbol               string   `protobuf:"bytes,1,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Mic                  string   `protobuf:"bytes,2,opt,name=mic,proto3" json:"mic,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ExactMatchParameters) Reset()         { *m = ExactMatchParameters{} }
func (m *ExactMatchParameters) String() string { return proto.CompactTextString(m) }
func (*ExactMatchParameters) ProtoMessage()    {}
func (*ExactMatchParameters) Descriptor() ([]byte, []int) {
	return fileDescriptor_bda44339ea58dbd5, []int{4}
}

func (m *ExactMatchParameters) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExactMatchParameters.Unmarshal(m, b)
}
func (m *ExactMatchParameters) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExactMatchParameters.Marshal(b, m, deterministic)
}
func (m *ExactMatchParameters) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExactMatchParameters.Merge(m, src)
}
func (m *ExactMatchParameters) XXX_Size() int {
	return xxx_messageInfo_ExactMatchParameters.Size(m)
}
func (m *ExactMatchParameters) XXX_DiscardUnknown() {
	xxx_messageInfo_ExactMatchParameters.DiscardUnknown(m)
}

var xxx_messageInfo_ExactMatchParameters proto.InternalMessageInfo

func (m *ExactMatchParameters) GetSymbol() string {
	if m != nil {
		return m.Symbol
	}
	return ""
}

func (m *ExactMatchParameters) GetMic() string {
	if m != nil {
		return m.Mic
	}
	return ""
}

func init() {
	proto.RegisterType((*ListingId)(nil), "staticdataservice.ListingId")
	proto.RegisterType((*ListingIds)(nil), "staticdataservice.ListingIds")
	proto.RegisterType((*Listings)(nil), "staticdataservice.Listings")
	proto.RegisterType((*MatchParameters)(nil), "staticdataservice.MatchParameters")
	proto.RegisterType((*ExactMatchParameters)(nil), "staticdataservice.ExactMatchParameters")
}

func init() { proto.RegisterFile("staticdataservice.proto", fileDescriptor_bda44339ea58dbd5) }

var fileDescriptor_bda44339ea58dbd5 = []byte{
	// 336 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0x4f, 0x4f, 0xc2, 0x40,
	0x10, 0xc5, 0x41, 0x02, 0x81, 0x21, 0xfe, 0x61, 0x34, 0x4a, 0x10, 0x0c, 0xd9, 0x8b, 0x68, 0x4c,
	0x0f, 0x98, 0x78, 0xf2, 0xe0, 0x41, 0x63, 0x88, 0x1a, 0xb5, 0x1c, 0xf4, 0xe0, 0x65, 0x29, 0x1b,
	0xd8, 0xa4, 0xdb, 0x9a, 0xee, 0x68, 0xf4, 0x13, 0xf9, 0x35, 0x0d, 0xbb, 0x4b, 0xdb, 0x00, 0xca,
	0xad, 0xfb, 0xde, 0xdb, 0x5f, 0xde, 0x4c, 0x17, 0x0e, 0x34, 0x71, 0x92, 0xc1, 0x98, 0x13, 0xd7,
	0x22, 0xf9, 0x94, 0x81, 0xf0, 0xde, 0x93, 0x98, 0x62, 0x6c, 0x2c, 0x19, 0xad, 0xcd, 0x50, 0x6a,
	0x92, 0xd1, 0xc4, 0x26, 0xd8, 0x09, 0xd4, 0xee, 0xad, 0x30, 0x18, 0x63, 0x1b, 0x6a, 0xe1, 0xfc,
	0xd0, 0x2c, 0x76, 0x8b, 0xbd, 0xb2, 0x9f, 0x09, 0xec, 0x0c, 0x20, 0x8d, 0x6a, 0x3c, 0x02, 0x48,
	0x2d, 0xdd, 0x2c, 0x76, 0x4b, 0xbd, 0xb2, 0x9f, 0x53, 0xd8, 0x05, 0x54, 0x5d, 0x5a, 0xe3, 0x29,
	0x54, 0x9d, 0x63, 0x93, 0xf5, 0xfe, 0x96, 0xa7, 0xe2, 0xb1, 0x08, 0x3d, 0x17, 0xf1, 0x53, 0x9f,
	0x3d, 0xc3, 0xf6, 0x03, 0xa7, 0x60, 0xfa, 0xc4, 0x13, 0xae, 0x04, 0x89, 0x44, 0x63, 0x17, 0xea,
	0xfa, 0x5b, 0x8d, 0xe2, 0xd0, 0x18, 0xa6, 0x58, 0xcd, 0xcf, 0x4b, 0xb3, 0xe2, 0x11, 0x57, 0xc2,
	0xfa, 0x1b, 0xc6, 0xcf, 0x04, 0x76, 0x05, 0x7b, 0x37, 0x5f, 0x3c, 0xa0, 0x45, 0xee, 0x3e, 0x54,
	0x2c, 0xc4, 0x21, 0xdd, 0x09, 0x77, 0xa0, 0xa4, 0x64, 0xe0, 0x38, 0xb3, 0xcf, 0xfe, 0x4f, 0x09,
	0x1a, 0x43, 0xb3, 0xca, 0x6b, 0x4e, 0x7c, 0x68, 0x57, 0x89, 0x6f, 0xd0, 0xb9, 0x15, 0x34, 0x9f,
	0xf2, 0x45, 0xd2, 0x74, 0xc8, 0x95, 0x18, 0x44, 0x9a, 0x92, 0x0f, 0x25, 0x22, 0xc2, 0xb6, 0xb7,
	0xfc, 0x63, 0xd2, 0x15, 0xb6, 0x0e, 0xff, 0x76, 0x35, 0x2b, 0xe0, 0x23, 0x60, 0x46, 0x37, 0xd5,
	0x65, 0x34, 0xc1, 0xe3, 0x15, 0x97, 0x56, 0x0d, 0xd7, 0x5a, 0xd8, 0x30, 0x2b, 0xe0, 0x2b, 0xec,
	0xe6, 0xea, 0xa6, 0x44, 0xb6, 0x82, 0xb8, 0x08, 0x5b, 0x53, 0xf5, 0x12, 0x20, 0x23, 0xaf, 0x99,
	0x7a, 0xb9, 0xd7, 0x1d, 0xd4, 0x73, 0xbd, 0xb0, 0xf3, 0xdf, 0xf5, 0x75, 0x55, 0x46, 0x15, 0xf3,
	0xac, 0xcf, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x89, 0xe6, 0x30, 0x7c, 0x13, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// StaticDataServiceClient is the client API for StaticDataService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type StaticDataServiceClient interface {
	GetListingsWithSameInstrument(ctx context.Context, in *ListingId, opts ...grpc.CallOption) (*Listings, error)
	GetListingMatching(ctx context.Context, in *ExactMatchParameters, opts ...grpc.CallOption) (*model.Listing, error)
	GetListingsMatching(ctx context.Context, in *MatchParameters, opts ...grpc.CallOption) (*Listings, error)
	GetListing(ctx context.Context, in *ListingId, opts ...grpc.CallOption) (*model.Listing, error)
	GetListings(ctx context.Context, in *ListingIds, opts ...grpc.CallOption) (*Listings, error)
}

type staticDataServiceClient struct {
	cc *grpc.ClientConn
}

func NewStaticDataServiceClient(cc *grpc.ClientConn) StaticDataServiceClient {
	return &staticDataServiceClient{cc}
}

func (c *staticDataServiceClient) GetListingsWithSameInstrument(ctx context.Context, in *ListingId, opts ...grpc.CallOption) (*Listings, error) {
	out := new(Listings)
	err := c.cc.Invoke(ctx, "/staticdataservice.StaticDataService/GetListingsWithSameInstrument", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *staticDataServiceClient) GetListingMatching(ctx context.Context, in *ExactMatchParameters, opts ...grpc.CallOption) (*model.Listing, error) {
	out := new(model.Listing)
	err := c.cc.Invoke(ctx, "/staticdataservice.StaticDataService/GetListingMatching", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *staticDataServiceClient) GetListingsMatching(ctx context.Context, in *MatchParameters, opts ...grpc.CallOption) (*Listings, error) {
	out := new(Listings)
	err := c.cc.Invoke(ctx, "/staticdataservice.StaticDataService/GetListingsMatching", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *staticDataServiceClient) GetListing(ctx context.Context, in *ListingId, opts ...grpc.CallOption) (*model.Listing, error) {
	out := new(model.Listing)
	err := c.cc.Invoke(ctx, "/staticdataservice.StaticDataService/GetListing", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *staticDataServiceClient) GetListings(ctx context.Context, in *ListingIds, opts ...grpc.CallOption) (*Listings, error) {
	out := new(Listings)
	err := c.cc.Invoke(ctx, "/staticdataservice.StaticDataService/GetListings", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StaticDataServiceServer is the server API for StaticDataService service.
type StaticDataServiceServer interface {
	GetListingsWithSameInstrument(context.Context, *ListingId) (*Listings, error)
	GetListingMatching(context.Context, *ExactMatchParameters) (*model.Listing, error)
	GetListingsMatching(context.Context, *MatchParameters) (*Listings, error)
	GetListing(context.Context, *ListingId) (*model.Listing, error)
	GetListings(context.Context, *ListingIds) (*Listings, error)
}

// UnimplementedStaticDataServiceServer can be embedded to have forward compatible implementations.
type UnimplementedStaticDataServiceServer struct {
}

func (*UnimplementedStaticDataServiceServer) GetListingsWithSameInstrument(ctx context.Context, req *ListingId) (*Listings, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetListingsWithSameInstrument not implemented")
}
func (*UnimplementedStaticDataServiceServer) GetListingMatching(ctx context.Context, req *ExactMatchParameters) (*model.Listing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetListingMatching not implemented")
}
func (*UnimplementedStaticDataServiceServer) GetListingsMatching(ctx context.Context, req *MatchParameters) (*Listings, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetListingsMatching not implemented")
}
func (*UnimplementedStaticDataServiceServer) GetListing(ctx context.Context, req *ListingId) (*model.Listing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetListing not implemented")
}
func (*UnimplementedStaticDataServiceServer) GetListings(ctx context.Context, req *ListingIds) (*Listings, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetListings not implemented")
}

func RegisterStaticDataServiceServer(s *grpc.Server, srv StaticDataServiceServer) {
	s.RegisterService(&_StaticDataService_serviceDesc, srv)
}

func _StaticDataService_GetListingsWithSameInstrument_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListingId)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StaticDataServiceServer).GetListingsWithSameInstrument(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/staticdataservice.StaticDataService/GetListingsWithSameInstrument",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StaticDataServiceServer).GetListingsWithSameInstrument(ctx, req.(*ListingId))
	}
	return interceptor(ctx, in, info, handler)
}

func _StaticDataService_GetListingMatching_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExactMatchParameters)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StaticDataServiceServer).GetListingMatching(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/staticdataservice.StaticDataService/GetListingMatching",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StaticDataServiceServer).GetListingMatching(ctx, req.(*ExactMatchParameters))
	}
	return interceptor(ctx, in, info, handler)
}

func _StaticDataService_GetListingsMatching_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MatchParameters)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StaticDataServiceServer).GetListingsMatching(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/staticdataservice.StaticDataService/GetListingsMatching",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StaticDataServiceServer).GetListingsMatching(ctx, req.(*MatchParameters))
	}
	return interceptor(ctx, in, info, handler)
}

func _StaticDataService_GetListing_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListingId)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StaticDataServiceServer).GetListing(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/staticdataservice.StaticDataService/GetListing",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StaticDataServiceServer).GetListing(ctx, req.(*ListingId))
	}
	return interceptor(ctx, in, info, handler)
}

func _StaticDataService_GetListings_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListingIds)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StaticDataServiceServer).GetListings(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/staticdataservice.StaticDataService/GetListings",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StaticDataServiceServer).GetListings(ctx, req.(*ListingIds))
	}
	return interceptor(ctx, in, info, handler)
}

var _StaticDataService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "staticdataservice.StaticDataService",
	HandlerType: (*StaticDataServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetListingsWithSameInstrument",
			Handler:    _StaticDataService_GetListingsWithSameInstrument_Handler,
		},
		{
			MethodName: "GetListingMatching",
			Handler:    _StaticDataService_GetListingMatching_Handler,
		},
		{
			MethodName: "GetListingsMatching",
			Handler:    _StaticDataService_GetListingsMatching_Handler,
		},
		{
			MethodName: "GetListing",
			Handler:    _StaticDataService_GetListing_Handler,
		},
		{
			MethodName: "GetListings",
			Handler:    _StaticDataService_GetListings_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "staticdataservice.proto",
}
