// axiomverse.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.29.1
// source: proto/keys.proto

package pb

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

type Algorithm int32

const (
	Algorithm_NONE         Algorithm = 0
	Algorithm_KYBER512     Algorithm = 1
	Algorithm_KYBER768     Algorithm = 2
	Algorithm_KYBER1024    Algorithm = 3
	Algorithm_FALCON512    Algorithm = 4
	Algorithm_DILITHIUM2   Algorithm = 5
	Algorithm_DILITHIUM3   Algorithm = 6
	Algorithm_EDWARDS25519 Algorithm = 7
	Algorithm_ECDSA        Algorithm = 8
	Algorithm_RSA          Algorithm = 9
	Algorithm_EDDSA        Algorithm = 10
)

// Enum value maps for Algorithm.
var (
	Algorithm_name = map[int32]string{
		0:  "NONE",
		1:  "KYBER512",
		2:  "KYBER768",
		3:  "KYBER1024",
		4:  "FALCON512",
		5:  "DILITHIUM2",
		6:  "DILITHIUM3",
		7:  "EDWARDS25519",
		8:  "ECDSA",
		9:  "RSA",
		10: "EDDSA",
	}
	Algorithm_value = map[string]int32{
		"NONE":         0,
		"KYBER512":     1,
		"KYBER768":     2,
		"KYBER1024":    3,
		"FALCON512":    4,
		"DILITHIUM2":   5,
		"DILITHIUM3":   6,
		"EDWARDS25519": 7,
		"ECDSA":        8,
		"RSA":          9,
		"EDDSA":        10,
	}
)

func (x Algorithm) Enum() *Algorithm {
	p := new(Algorithm)
	*p = x
	return p
}

func (x Algorithm) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Algorithm) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_keys_proto_enumTypes[0].Descriptor()
}

func (Algorithm) Type() protoreflect.EnumType {
	return &file_proto_keys_proto_enumTypes[0]
}

func (x Algorithm) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Algorithm.Descriptor instead.
func (Algorithm) EnumDescriptor() ([]byte, []int) {
	return file_proto_keys_proto_rawDescGZIP(), []int{0}
}

type Key struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Algorithm Algorithm `protobuf:"varint,1,opt,name=algorithm,proto3,enum=pb.Algorithm" json:"algorithm,omitempty"`
	Keys      string    `protobuf:"bytes,2,opt,name=keys,proto3" json:"keys,omitempty"`
}

func (x *Key) Reset() {
	*x = Key{}
	mi := &file_proto_keys_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Key) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Key) ProtoMessage() {}

func (x *Key) ProtoReflect() protoreflect.Message {
	mi := &file_proto_keys_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Key.ProtoReflect.Descriptor instead.
func (*Key) Descriptor() ([]byte, []int) {
	return file_proto_keys_proto_rawDescGZIP(), []int{0}
}

func (x *Key) GetAlgorithm() Algorithm {
	if x != nil {
		return x.Algorithm
	}
	return Algorithm_NONE
}

func (x *Key) GetKeys() string {
	if x != nil {
		return x.Keys
	}
	return ""
}

var File_proto_keys_proto protoreflect.FileDescriptor

var file_proto_keys_proto_rawDesc = []byte{
	0x0a, 0x10, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6b, 0x65, 0x79, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x22, 0x46, 0x0a, 0x03, 0x4b, 0x65, 0x79, 0x12, 0x2b, 0x0a,
	0x09, 0x61, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x0d, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x52,
	0x09, 0x61, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x65,
	0x79, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x2a, 0xa0,
	0x01, 0x0a, 0x09, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x12, 0x08, 0x0a, 0x04,
	0x4e, 0x4f, 0x4e, 0x45, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x4b, 0x59, 0x42, 0x45, 0x52, 0x35,
	0x31, 0x32, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x4b, 0x59, 0x42, 0x45, 0x52, 0x37, 0x36, 0x38,
	0x10, 0x02, 0x12, 0x0d, 0x0a, 0x09, 0x4b, 0x59, 0x42, 0x45, 0x52, 0x31, 0x30, 0x32, 0x34, 0x10,
	0x03, 0x12, 0x0d, 0x0a, 0x09, 0x46, 0x41, 0x4c, 0x43, 0x4f, 0x4e, 0x35, 0x31, 0x32, 0x10, 0x04,
	0x12, 0x0e, 0x0a, 0x0a, 0x44, 0x49, 0x4c, 0x49, 0x54, 0x48, 0x49, 0x55, 0x4d, 0x32, 0x10, 0x05,
	0x12, 0x0e, 0x0a, 0x0a, 0x44, 0x49, 0x4c, 0x49, 0x54, 0x48, 0x49, 0x55, 0x4d, 0x33, 0x10, 0x06,
	0x12, 0x10, 0x0a, 0x0c, 0x45, 0x44, 0x57, 0x41, 0x52, 0x44, 0x53, 0x32, 0x35, 0x35, 0x31, 0x39,
	0x10, 0x07, 0x12, 0x09, 0x0a, 0x05, 0x45, 0x43, 0x44, 0x53, 0x41, 0x10, 0x08, 0x12, 0x07, 0x0a,
	0x03, 0x52, 0x53, 0x41, 0x10, 0x09, 0x12, 0x09, 0x0a, 0x05, 0x45, 0x44, 0x44, 0x53, 0x41, 0x10,
	0x0a, 0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x74, 0x68, 0x65, 0x61, 0x78, 0x69, 0x6f, 0x6d, 0x76, 0x65, 0x72, 0x73, 0x65, 0x2f, 0x68, 0x79,
	0x64, 0x61, 0x70, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x3b, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_keys_proto_rawDescOnce sync.Once
	file_proto_keys_proto_rawDescData = file_proto_keys_proto_rawDesc
)

func file_proto_keys_proto_rawDescGZIP() []byte {
	file_proto_keys_proto_rawDescOnce.Do(func() {
		file_proto_keys_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_keys_proto_rawDescData)
	})
	return file_proto_keys_proto_rawDescData
}

var file_proto_keys_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_keys_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_proto_keys_proto_goTypes = []any{
	(Algorithm)(0), // 0: pb.Algorithm
	(*Key)(nil),    // 1: pb.Key
}
var file_proto_keys_proto_depIdxs = []int32{
	0, // 0: pb.Key.algorithm:type_name -> pb.Algorithm
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proto_keys_proto_init() }
func file_proto_keys_proto_init() {
	if File_proto_keys_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_keys_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_keys_proto_goTypes,
		DependencyIndexes: file_proto_keys_proto_depIdxs,
		EnumInfos:         file_proto_keys_proto_enumTypes,
		MessageInfos:      file_proto_keys_proto_msgTypes,
	}.Build()
	File_proto_keys_proto = out.File
	file_proto_keys_proto_rawDesc = nil
	file_proto_keys_proto_goTypes = nil
	file_proto_keys_proto_depIdxs = nil
}
