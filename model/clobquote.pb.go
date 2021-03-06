// Code generated by protoc-gen-go. DO NOT EDIT.
// source: clobquote.proto

package model

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type ClobLine struct {
	Size                 *Decimal64 `protobuf:"bytes,1,opt,name=size,proto3" json:"size,omitempty"`
	Price                *Decimal64 `protobuf:"bytes,2,opt,name=price,proto3" json:"price,omitempty"`
	EntryId              string     `protobuf:"bytes,3,opt,name=entryId,proto3" json:"entryId,omitempty"`
	ListingId            int32      `protobuf:"varint,4,opt,name=listingId,proto3" json:"listingId,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *ClobLine) Reset()         { *m = ClobLine{} }
func (m *ClobLine) String() string { return proto.CompactTextString(m) }
func (*ClobLine) ProtoMessage()    {}
func (*ClobLine) Descriptor() ([]byte, []int) {
	return fileDescriptor_eff833333d312bfe, []int{0}
}

func (m *ClobLine) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ClobLine.Unmarshal(m, b)
}
func (m *ClobLine) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ClobLine.Marshal(b, m, deterministic)
}
func (m *ClobLine) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ClobLine.Merge(m, src)
}
func (m *ClobLine) XXX_Size() int {
	return xxx_messageInfo_ClobLine.Size(m)
}
func (m *ClobLine) XXX_DiscardUnknown() {
	xxx_messageInfo_ClobLine.DiscardUnknown(m)
}

var xxx_messageInfo_ClobLine proto.InternalMessageInfo

func (m *ClobLine) GetSize() *Decimal64 {
	if m != nil {
		return m.Size
	}
	return nil
}

func (m *ClobLine) GetPrice() *Decimal64 {
	if m != nil {
		return m.Price
	}
	return nil
}

func (m *ClobLine) GetEntryId() string {
	if m != nil {
		return m.EntryId
	}
	return ""
}

func (m *ClobLine) GetListingId() int32 {
	if m != nil {
		return m.ListingId
	}
	return 0
}

type ClobQuote struct {
	ListingId            int32       `protobuf:"varint,1,opt,name=listingId,proto3" json:"listingId,omitempty"`
	Bids                 []*ClobLine `protobuf:"bytes,2,rep,name=bids,proto3" json:"bids,omitempty"`
	Offers               []*ClobLine `protobuf:"bytes,3,rep,name=offers,proto3" json:"offers,omitempty"`
	StreamInterrupted    bool        `protobuf:"varint,4,opt,name=streamInterrupted,proto3" json:"streamInterrupted,omitempty"`
	StreamStatusMsg      string      `protobuf:"bytes,5,opt,name=streamStatusMsg,proto3" json:"streamStatusMsg,omitempty"`
	LastPrice            *Decimal64  `protobuf:"bytes,6,opt,name=lastPrice,proto3" json:"lastPrice,omitempty"`
	LastQuantity         *Decimal64  `protobuf:"bytes,7,opt,name=lastQuantity,proto3" json:"lastQuantity,omitempty"`
	TradedVolume         *Decimal64  `protobuf:"bytes,8,opt,name=tradedVolume,proto3" json:"tradedVolume,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *ClobQuote) Reset()         { *m = ClobQuote{} }
func (m *ClobQuote) String() string { return proto.CompactTextString(m) }
func (*ClobQuote) ProtoMessage()    {}
func (*ClobQuote) Descriptor() ([]byte, []int) {
	return fileDescriptor_eff833333d312bfe, []int{1}
}

func (m *ClobQuote) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ClobQuote.Unmarshal(m, b)
}
func (m *ClobQuote) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ClobQuote.Marshal(b, m, deterministic)
}
func (m *ClobQuote) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ClobQuote.Merge(m, src)
}
func (m *ClobQuote) XXX_Size() int {
	return xxx_messageInfo_ClobQuote.Size(m)
}
func (m *ClobQuote) XXX_DiscardUnknown() {
	xxx_messageInfo_ClobQuote.DiscardUnknown(m)
}

var xxx_messageInfo_ClobQuote proto.InternalMessageInfo

func (m *ClobQuote) GetListingId() int32 {
	if m != nil {
		return m.ListingId
	}
	return 0
}

func (m *ClobQuote) GetBids() []*ClobLine {
	if m != nil {
		return m.Bids
	}
	return nil
}

func (m *ClobQuote) GetOffers() []*ClobLine {
	if m != nil {
		return m.Offers
	}
	return nil
}

func (m *ClobQuote) GetStreamInterrupted() bool {
	if m != nil {
		return m.StreamInterrupted
	}
	return false
}

func (m *ClobQuote) GetStreamStatusMsg() string {
	if m != nil {
		return m.StreamStatusMsg
	}
	return ""
}

func (m *ClobQuote) GetLastPrice() *Decimal64 {
	if m != nil {
		return m.LastPrice
	}
	return nil
}

func (m *ClobQuote) GetLastQuantity() *Decimal64 {
	if m != nil {
		return m.LastQuantity
	}
	return nil
}

func (m *ClobQuote) GetTradedVolume() *Decimal64 {
	if m != nil {
		return m.TradedVolume
	}
	return nil
}

func init() {
	proto.RegisterType((*ClobLine)(nil), "model.ClobLine")
	proto.RegisterType((*ClobQuote)(nil), "model.ClobQuote")
}

func init() { proto.RegisterFile("clobquote.proto", fileDescriptor_eff833333d312bfe) }

var fileDescriptor_eff833333d312bfe = []byte{
	// 317 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0xcf, 0x4a, 0xc3, 0x40,
	0x10, 0x87, 0x49, 0xdb, 0xb4, 0xcd, 0x54, 0xa8, 0xdd, 0xd3, 0x22, 0x1e, 0x42, 0x15, 0xcd, 0x41,
	0x72, 0xa8, 0xc5, 0x07, 0x50, 0x2f, 0x05, 0x05, 0x1b, 0xc1, 0x83, 0xb7, 0xfc, 0x99, 0x96, 0x85,
	0x4d, 0x36, 0xee, 0x4e, 0x0e, 0xf5, 0x29, 0x7c, 0x4f, 0x5f, 0x42, 0xb2, 0x69, 0xa9, 0xad, 0xe6,
	0xb4, 0xec, 0x7c, 0xdf, 0xc0, 0xfc, 0x86, 0x81, 0x71, 0x2a, 0x55, 0xf2, 0x51, 0x29, 0xc2, 0xb0,
	0xd4, 0x8a, 0x14, 0x73, 0x73, 0x95, 0xa1, 0x3c, 0x9b, 0xd8, 0x27, 0x55, 0x79, 0xae, 0x8a, 0x86,
	0x4c, 0xbf, 0x1c, 0x18, 0x3e, 0x48, 0x95, 0x3c, 0x89, 0x02, 0xd9, 0x25, 0xf4, 0x8c, 0xf8, 0x44,
	0xee, 0xf8, 0x4e, 0x30, 0x9a, 0x9d, 0x86, 0x56, 0x0f, 0x1f, 0x31, 0x15, 0x79, 0x2c, 0xef, 0xe6,
	0x91, 0xa5, 0xec, 0x0a, 0xdc, 0x52, 0x8b, 0x14, 0x79, 0xa7, 0x45, 0x6b, 0x30, 0xe3, 0x30, 0xc0,
	0x82, 0xf4, 0x66, 0x91, 0xf1, 0xae, 0xef, 0x04, 0x5e, 0xb4, 0xfb, 0xb2, 0x73, 0xf0, 0xa4, 0x30,
	0x24, 0x8a, 0xf5, 0x22, 0xe3, 0x3d, 0xdf, 0x09, 0xdc, 0x68, 0x5f, 0x98, 0x7e, 0x77, 0xc0, 0xab,
	0x47, 0x5a, 0xd6, 0x01, 0x0e, 0x5d, 0xe7, 0xc8, 0x65, 0x17, 0xd0, 0x4b, 0x44, 0x66, 0x78, 0xc7,
	0xef, 0x06, 0xa3, 0xd9, 0x78, 0x3b, 0xca, 0x2e, 0x50, 0x64, 0x21, 0xbb, 0x86, 0xbe, 0x5a, 0xad,
	0x50, 0x1b, 0xde, 0xfd, 0x5f, 0xdb, 0x62, 0x76, 0x03, 0x13, 0x43, 0x1a, 0xe3, 0x7c, 0x51, 0x10,
	0x6a, 0x5d, 0x95, 0x84, 0xcd, 0x7c, 0xc3, 0xe8, 0x2f, 0x60, 0x01, 0x8c, 0x9b, 0xe2, 0x2b, 0xc5,
	0x54, 0x99, 0x67, 0xb3, 0xe6, 0xae, 0xcd, 0x79, 0x5c, 0x66, 0x21, 0x78, 0x32, 0x36, 0xf4, 0x62,
	0xb7, 0xd6, 0x6f, 0xd9, 0xda, 0x5e, 0x61, 0x73, 0x38, 0xa9, 0x3f, 0xcb, 0x2a, 0x2e, 0x48, 0xd0,
	0x86, 0x0f, 0x5a, 0x5a, 0x0e, 0xac, 0xba, 0x8b, 0x74, 0x9c, 0x61, 0xf6, 0xa6, 0x64, 0x95, 0x23,
	0x1f, 0xb6, 0x75, 0xfd, 0xb6, 0xee, 0x07, 0xef, 0xcd, 0x71, 0x24, 0x7d, 0x7b, 0x10, 0xb7, 0x3f,
	0x01, 0x00, 0x00, 0xff, 0xff, 0x21, 0xd1, 0x5c, 0xf2, 0x3d, 0x02, 0x00, 0x00,
}
