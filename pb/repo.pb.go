// Code generated by protoc-gen-go. DO NOT EDIT.
// source: repo.proto

package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _struct "github.com/golang/protobuf/ptypes/struct"
import timestamp "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type File struct {
	Mill                 string               `protobuf:"bytes,1,opt,name=mill,proto3" json:"mill,omitempty"`
	Checksum             string               `protobuf:"bytes,2,opt,name=checksum,proto3" json:"checksum,omitempty"`
	Source               string               `protobuf:"bytes,3,opt,name=source,proto3" json:"source,omitempty"`
	Opts                 string               `protobuf:"bytes,4,opt,name=opts,proto3" json:"opts,omitempty"`
	Hash                 string               `protobuf:"bytes,5,opt,name=hash,proto3" json:"hash,omitempty"`
	Key                  string               `protobuf:"bytes,6,opt,name=key,proto3" json:"key,omitempty"`
	Media                string               `protobuf:"bytes,7,opt,name=media,proto3" json:"media,omitempty"`
	Name                 string               `protobuf:"bytes,8,opt,name=name,proto3" json:"name,omitempty"`
	Size                 int64                `protobuf:"varint,9,opt,name=size,proto3" json:"size,omitempty"`
	Added                *timestamp.Timestamp `protobuf:"bytes,10,opt,name=added,proto3" json:"added,omitempty"`
	Meta                 *_struct.Struct      `protobuf:"bytes,11,opt,name=meta,proto3" json:"meta,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *File) Reset()         { *m = File{} }
func (m *File) String() string { return proto.CompactTextString(m) }
func (*File) ProtoMessage()    {}
func (*File) Descriptor() ([]byte, []int) {
	return fileDescriptor_repo_0d9354398edfe73b, []int{0}
}
func (m *File) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_File.Unmarshal(m, b)
}
func (m *File) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_File.Marshal(b, m, deterministic)
}
func (dst *File) XXX_Merge(src proto.Message) {
	xxx_messageInfo_File.Merge(dst, src)
}
func (m *File) XXX_Size() int {
	return xxx_messageInfo_File.Size(m)
}
func (m *File) XXX_DiscardUnknown() {
	xxx_messageInfo_File.DiscardUnknown(m)
}

var xxx_messageInfo_File proto.InternalMessageInfo

func (m *File) GetMill() string {
	if m != nil {
		return m.Mill
	}
	return ""
}

func (m *File) GetChecksum() string {
	if m != nil {
		return m.Checksum
	}
	return ""
}

func (m *File) GetSource() string {
	if m != nil {
		return m.Source
	}
	return ""
}

func (m *File) GetOpts() string {
	if m != nil {
		return m.Opts
	}
	return ""
}

func (m *File) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *File) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *File) GetMedia() string {
	if m != nil {
		return m.Media
	}
	return ""
}

func (m *File) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *File) GetSize() int64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *File) GetAdded() *timestamp.Timestamp {
	if m != nil {
		return m.Added
	}
	return nil
}

func (m *File) GetMeta() *_struct.Struct {
	if m != nil {
		return m.Meta
	}
	return nil
}

type Directory struct {
	Files                map[string]*File `protobuf:"bytes,1,rep,name=files,proto3" json:"files,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *Directory) Reset()         { *m = Directory{} }
func (m *Directory) String() string { return proto.CompactTextString(m) }
func (*Directory) ProtoMessage()    {}
func (*Directory) Descriptor() ([]byte, []int) {
	return fileDescriptor_repo_0d9354398edfe73b, []int{1}
}
func (m *Directory) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Directory.Unmarshal(m, b)
}
func (m *Directory) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Directory.Marshal(b, m, deterministic)
}
func (dst *Directory) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Directory.Merge(dst, src)
}
func (m *Directory) XXX_Size() int {
	return xxx_messageInfo_Directory.Size(m)
}
func (m *Directory) XXX_DiscardUnknown() {
	xxx_messageInfo_Directory.DiscardUnknown(m)
}

var xxx_messageInfo_Directory proto.InternalMessageInfo

func (m *Directory) GetFiles() map[string]*File {
	if m != nil {
		return m.Files
	}
	return nil
}

func init() {
	proto.RegisterType((*File)(nil), "File")
	proto.RegisterType((*Directory)(nil), "Directory")
	proto.RegisterMapType((map[string]*File)(nil), "Directory.FilesEntry")
}

func init() { proto.RegisterFile("repo.proto", fileDescriptor_repo_0d9354398edfe73b) }

var fileDescriptor_repo_0d9354398edfe73b = []byte{
	// 325 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0xcb, 0x6a, 0xf3, 0x30,
	0x10, 0x85, 0xf1, 0x2d, 0x7f, 0x32, 0xde, 0xfc, 0x88, 0x5e, 0x84, 0x5b, 0xa8, 0xc9, 0xca, 0x10,
	0x70, 0x4a, 0xba, 0x29, 0xdd, 0x14, 0x4a, 0xdb, 0x07, 0x70, 0xbb, 0xea, 0x4e, 0xb1, 0x27, 0x89,
	0x88, 0x15, 0x19, 0x49, 0x2e, 0xb8, 0xef, 0xd3, 0xf7, 0x2c, 0x92, 0x72, 0x81, 0x76, 0x77, 0xe6,
	0xd3, 0x99, 0x41, 0xe7, 0x00, 0x28, 0xec, 0x64, 0xd9, 0x29, 0x69, 0x64, 0x76, 0xb3, 0x96, 0x72,
	0xdd, 0xe2, 0xdc, 0x4d, 0xcb, 0x7e, 0x35, 0x37, 0x5c, 0xa0, 0x36, 0x4c, 0x74, 0x7b, 0xc3, 0xf5,
	0x6f, 0x83, 0x36, 0xaa, 0xaf, 0x8d, 0x7f, 0x9d, 0x7e, 0x87, 0x10, 0xbf, 0xf2, 0x16, 0x09, 0x81,
	0x58, 0xf0, 0xb6, 0xa5, 0x41, 0x1e, 0x14, 0x93, 0xca, 0x69, 0x92, 0xc1, 0xb8, 0xde, 0x60, 0xbd,
	0xd5, 0xbd, 0xa0, 0xa1, 0xe3, 0xc7, 0x99, 0x5c, 0xc0, 0x48, 0xcb, 0x5e, 0xd5, 0x48, 0x23, 0xf7,
	0xb2, 0x9f, 0xec, 0x1d, 0xd9, 0x19, 0x4d, 0x63, 0x7f, 0xc7, 0x6a, 0xcb, 0x36, 0x4c, 0x6f, 0x68,
	0xe2, 0x99, 0xd5, 0xe4, 0x3f, 0x44, 0x5b, 0x1c, 0xe8, 0xc8, 0x21, 0x2b, 0xc9, 0x19, 0x24, 0x02,
	0x1b, 0xce, 0xe8, 0x3f, 0xc7, 0xfc, 0x60, 0x77, 0x77, 0x4c, 0x20, 0x1d, 0xfb, 0x5d, 0xab, 0x2d,
	0xd3, 0xfc, 0x0b, 0xe9, 0x24, 0x0f, 0x8a, 0xa8, 0x72, 0x9a, 0xdc, 0x42, 0xc2, 0x9a, 0x06, 0x1b,
	0x0a, 0x79, 0x50, 0xa4, 0x8b, 0xac, 0xf4, 0xb1, 0xcb, 0x43, 0xec, 0xf2, 0xfd, 0xd0, 0x4b, 0xe5,
	0x8d, 0x64, 0x06, 0xb1, 0x40, 0xc3, 0x68, 0xea, 0x16, 0x2e, 0xff, 0x2c, 0xbc, 0xb9, 0x9e, 0x2a,
	0x67, 0x9a, 0x0e, 0x30, 0x79, 0xe6, 0x0a, 0x6b, 0x23, 0xd5, 0x40, 0x66, 0x90, 0xac, 0x78, 0x8b,
	0x9a, 0x06, 0x79, 0x54, 0xa4, 0x8b, 0xf3, 0xf2, 0xf8, 0x54, 0xda, 0x2e, 0xf5, 0xcb, 0xce, 0xa8,
	0xa1, 0xf2, 0x9e, 0xec, 0x11, 0xe0, 0x04, 0x0f, 0xb1, 0x83, 0x53, 0xec, 0x2b, 0x48, 0x3e, 0x59,
	0xdb, 0xa3, 0x6b, 0x38, 0x5d, 0x24, 0xee, 0x44, 0xe5, 0xd9, 0x43, 0x78, 0x1f, 0x3c, 0xc5, 0x1f,
	0x61, 0xb7, 0x5c, 0x8e, 0xdc, 0xbf, 0xee, 0x7e, 0x02, 0x00, 0x00, 0xff, 0xff, 0x99, 0x57, 0x10,
	0xb7, 0xfc, 0x01, 0x00, 0x00,
}