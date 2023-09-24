// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.0
// source: file_transfer/FileTransfer.proto

package file_transfer

import (
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

type ErrorCodes int32

const (
	ErrorCodes_OK                   ErrorCodes = 0
	ErrorCodes_S3_ERROR             ErrorCodes = 21
	ErrorCodes_CHUNK_ERROR          ErrorCodes = 22
	ErrorCodes_INVALID_METADATA     ErrorCodes = 23
	ErrorCodes_FILE_CREATE_ERROR    ErrorCodes = 24
	ErrorCodes_FILE_NOT_FOUND_ERROR ErrorCodes = 25
	ErrorCodes_FILE_WRITE_ERROR     ErrorCodes = 26
	ErrorCodes_FILE_READ_ERROR      ErrorCodes = 27
	ErrorCodes_STREAM_CLOSE_ERROR   ErrorCodes = 28
	ErrorCodes_REQUEST_FAILED       ErrorCodes = 29
)

// Enum value maps for ErrorCodes.
var (
	ErrorCodes_name = map[int32]string{
		0:  "OK",
		21: "S3_ERROR",
		22: "CHUNK_ERROR",
		23: "INVALID_METADATA",
		24: "FILE_CREATE_ERROR",
		25: "FILE_NOT_FOUND_ERROR",
		26: "FILE_WRITE_ERROR",
		27: "FILE_READ_ERROR",
		28: "STREAM_CLOSE_ERROR",
		29: "REQUEST_FAILED",
	}
	ErrorCodes_value = map[string]int32{
		"OK":                   0,
		"S3_ERROR":             21,
		"CHUNK_ERROR":          22,
		"INVALID_METADATA":     23,
		"FILE_CREATE_ERROR":    24,
		"FILE_NOT_FOUND_ERROR": 25,
		"FILE_WRITE_ERROR":     26,
		"FILE_READ_ERROR":      27,
		"STREAM_CLOSE_ERROR":   28,
		"REQUEST_FAILED":       29,
	}
)

func (x ErrorCodes) Enum() *ErrorCodes {
	p := new(ErrorCodes)
	*p = x
	return p
}

func (x ErrorCodes) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ErrorCodes) Descriptor() protoreflect.EnumDescriptor {
	return file_file_transfer_FileTransfer_proto_enumTypes[0].Descriptor()
}

func (ErrorCodes) Type() protoreflect.EnumType {
	return &file_file_transfer_FileTransfer_proto_enumTypes[0]
}

func (x ErrorCodes) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ErrorCodes.Descriptor instead.
func (ErrorCodes) EnumDescriptor() ([]byte, []int) {
	return file_file_transfer_FileTransfer_proto_rawDescGZIP(), []int{0}
}

type FileDownloadRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FileName  string `protobuf:"bytes,1,opt,name=fileName,proto3" json:"fileName,omitempty"`
	RequestId string `protobuf:"bytes,2,opt,name=requestId,proto3" json:"requestId,omitempty"`
}

func (x *FileDownloadRequest) Reset() {
	*x = FileDownloadRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_transfer_FileTransfer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileDownloadRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileDownloadRequest) ProtoMessage() {}

func (x *FileDownloadRequest) ProtoReflect() protoreflect.Message {
	mi := &file_file_transfer_FileTransfer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileDownloadRequest.ProtoReflect.Descriptor instead.
func (*FileDownloadRequest) Descriptor() ([]byte, []int) {
	return file_file_transfer_FileTransfer_proto_rawDescGZIP(), []int{0}
}

func (x *FileDownloadRequest) GetFileName() string {
	if x != nil {
		return x.FileName
	}
	return ""
}

func (x *FileDownloadRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

type FileChunk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Chunk []byte `protobuf:"bytes,1,opt,name=chunk,proto3" json:"chunk,omitempty"`
}

func (x *FileChunk) Reset() {
	*x = FileChunk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_transfer_FileTransfer_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileChunk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileChunk) ProtoMessage() {}

func (x *FileChunk) ProtoReflect() protoreflect.Message {
	mi := &file_file_transfer_FileTransfer_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileChunk.ProtoReflect.Descriptor instead.
func (*FileChunk) Descriptor() ([]byte, []int) {
	return file_file_transfer_FileTransfer_proto_rawDescGZIP(), []int{1}
}

func (x *FileChunk) GetChunk() []byte {
	if x != nil {
		return x.Chunk
	}
	return nil
}

type FileResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId string `protobuf:"bytes,1,opt,name=requestId,proto3" json:"requestId,omitempty"`
	Success   bool   `protobuf:"varint,2,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *FileResponse) Reset() {
	*x = FileResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_transfer_FileTransfer_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileResponse) ProtoMessage() {}

func (x *FileResponse) ProtoReflect() protoreflect.Message {
	mi := &file_file_transfer_FileTransfer_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileResponse.ProtoReflect.Descriptor instead.
func (*FileResponse) Descriptor() ([]byte, []int) {
	return file_file_transfer_FileTransfer_proto_rawDescGZIP(), []int{2}
}

func (x *FileResponse) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *FileResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type FileDeleteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FileName  string `protobuf:"bytes,1,opt,name=fileName,proto3" json:"fileName,omitempty"`
	RequestId string `protobuf:"bytes,2,opt,name=requestId,proto3" json:"requestId,omitempty"`
}

func (x *FileDeleteRequest) Reset() {
	*x = FileDeleteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_transfer_FileTransfer_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileDeleteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileDeleteRequest) ProtoMessage() {}

func (x *FileDeleteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_file_transfer_FileTransfer_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileDeleteRequest.ProtoReflect.Descriptor instead.
func (*FileDeleteRequest) Descriptor() ([]byte, []int) {
	return file_file_transfer_FileTransfer_proto_rawDescGZIP(), []int{3}
}

func (x *FileDeleteRequest) GetFileName() string {
	if x != nil {
		return x.FileName
	}
	return ""
}

func (x *FileDeleteRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

var File_file_transfer_FileTransfer_proto protoreflect.FileDescriptor

var file_file_transfer_FileTransfer_proto_rawDesc = []byte{
	0x0a, 0x20, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2f,
	0x46, 0x69, 0x6c, 0x65, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0d, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65,
	0x72, 0x22, 0x4f, 0x0a, 0x13, 0x46, 0x69, 0x6c, 0x65, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61,
	0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x49, 0x64, 0x22, 0x21, 0x0a, 0x09, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x12,
	0x14, 0x0a, 0x05, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05,
	0x63, 0x68, 0x75, 0x6e, 0x6b, 0x22, 0x46, 0x0a, 0x0c, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x22, 0x4d, 0x0a,
	0x11, 0x46, 0x69, 0x6c, 0x65, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1c,
	0x0a, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x2a, 0xd1, 0x01, 0x0a,
	0x0a, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x73, 0x12, 0x06, 0x0a, 0x02, 0x4f,
	0x4b, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x53, 0x33, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10,
	0x15, 0x12, 0x0f, 0x0a, 0x0b, 0x43, 0x48, 0x55, 0x4e, 0x4b, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52,
	0x10, 0x16, 0x12, 0x14, 0x0a, 0x10, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x4d, 0x45,
	0x54, 0x41, 0x44, 0x41, 0x54, 0x41, 0x10, 0x17, 0x12, 0x15, 0x0a, 0x11, 0x46, 0x49, 0x4c, 0x45,
	0x5f, 0x43, 0x52, 0x45, 0x41, 0x54, 0x45, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x18, 0x12,
	0x18, 0x0a, 0x14, 0x46, 0x49, 0x4c, 0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e,
	0x44, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x19, 0x12, 0x14, 0x0a, 0x10, 0x46, 0x49, 0x4c,
	0x45, 0x5f, 0x57, 0x52, 0x49, 0x54, 0x45, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x1a, 0x12,
	0x13, 0x0a, 0x0f, 0x46, 0x49, 0x4c, 0x45, 0x5f, 0x52, 0x45, 0x41, 0x44, 0x5f, 0x45, 0x52, 0x52,
	0x4f, 0x52, 0x10, 0x1b, 0x12, 0x16, 0x0a, 0x12, 0x53, 0x54, 0x52, 0x45, 0x41, 0x4d, 0x5f, 0x43,
	0x4c, 0x4f, 0x53, 0x45, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x1c, 0x12, 0x12, 0x0a, 0x0e,
	0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x5f, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x1d,
	0x32, 0xe5, 0x01, 0x0a, 0x0b, 0x46, 0x69, 0x6c, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x4a, 0x0a, 0x08, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x22, 0x2e, 0x66,
	0x69, 0x6c, 0x65, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x46, 0x69, 0x6c,
	0x65, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x18, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72,
	0x2e, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x30, 0x01, 0x12, 0x41, 0x0a, 0x06,
	0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x18, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x75, 0x6e, 0x6b,
	0x1a, 0x1b, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72,
	0x2e, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28, 0x01, 0x12,
	0x47, 0x0a, 0x06, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x20, 0x2e, 0x66, 0x69, 0x6c, 0x65,
	0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x66, 0x69,
	0x6c, 0x65, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x46, 0x69, 0x6c, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x65, 0x0a, 0x0f, 0x45, 0x64, 0x67, 0x65,
	0x46, 0x69, 0x6c, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x52, 0x0a, 0x10, 0x44,
	0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x46, 0x72, 0x6f, 0x6d, 0x45, 0x64, 0x67, 0x65, 0x12,
	0x22, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e,
	0x46, 0x69, 0x6c, 0x65, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x66, 0x65, 0x72, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x30, 0x01, 0x42,
	0x18, 0x5a, 0x16, 0x2e, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x69, 0x6c, 0x65,
	0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_file_transfer_FileTransfer_proto_rawDescOnce sync.Once
	file_file_transfer_FileTransfer_proto_rawDescData = file_file_transfer_FileTransfer_proto_rawDesc
)

func file_file_transfer_FileTransfer_proto_rawDescGZIP() []byte {
	file_file_transfer_FileTransfer_proto_rawDescOnce.Do(func() {
		file_file_transfer_FileTransfer_proto_rawDescData = protoimpl.X.CompressGZIP(file_file_transfer_FileTransfer_proto_rawDescData)
	})
	return file_file_transfer_FileTransfer_proto_rawDescData
}

var file_file_transfer_FileTransfer_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_file_transfer_FileTransfer_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_file_transfer_FileTransfer_proto_goTypes = []interface{}{
	(ErrorCodes)(0),             // 0: file_transfer.ErrorCodes
	(*FileDownloadRequest)(nil), // 1: file_transfer.FileDownloadRequest
	(*FileChunk)(nil),           // 2: file_transfer.FileChunk
	(*FileResponse)(nil),        // 3: file_transfer.FileResponse
	(*FileDeleteRequest)(nil),   // 4: file_transfer.FileDeleteRequest
}
var file_file_transfer_FileTransfer_proto_depIdxs = []int32{
	1, // 0: file_transfer.FileService.Download:input_type -> file_transfer.FileDownloadRequest
	2, // 1: file_transfer.FileService.Upload:input_type -> file_transfer.FileChunk
	4, // 2: file_transfer.FileService.Delete:input_type -> file_transfer.FileDeleteRequest
	1, // 3: file_transfer.EdgeFileService.DownloadFromEdge:input_type -> file_transfer.FileDownloadRequest
	2, // 4: file_transfer.FileService.Download:output_type -> file_transfer.FileChunk
	3, // 5: file_transfer.FileService.Upload:output_type -> file_transfer.FileResponse
	3, // 6: file_transfer.FileService.Delete:output_type -> file_transfer.FileResponse
	2, // 7: file_transfer.EdgeFileService.DownloadFromEdge:output_type -> file_transfer.FileChunk
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_file_transfer_FileTransfer_proto_init() }
func file_file_transfer_FileTransfer_proto_init() {
	if File_file_transfer_FileTransfer_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_file_transfer_FileTransfer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileDownloadRequest); i {
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
		file_file_transfer_FileTransfer_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileChunk); i {
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
		file_file_transfer_FileTransfer_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileResponse); i {
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
		file_file_transfer_FileTransfer_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileDeleteRequest); i {
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
			RawDescriptor: file_file_transfer_FileTransfer_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_file_transfer_FileTransfer_proto_goTypes,
		DependencyIndexes: file_file_transfer_FileTransfer_proto_depIdxs,
		EnumInfos:         file_file_transfer_FileTransfer_proto_enumTypes,
		MessageInfos:      file_file_transfer_FileTransfer_proto_msgTypes,
	}.Build()
	File_file_transfer_FileTransfer_proto = out.File
	file_file_transfer_FileTransfer_proto_rawDesc = nil
	file_file_transfer_FileTransfer_proto_goTypes = nil
	file_file_transfer_FileTransfer_proto_depIdxs = nil
}
