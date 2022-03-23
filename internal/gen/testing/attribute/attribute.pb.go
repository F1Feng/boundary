// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: testing/attribute/v1/attribute.proto

package attribute

import (
	_ "github.com/hashicorp/boundary/sdk/pbs/controller/protooptions"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TestResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Types that are assignable to Attrs:
	//	*TestResource_Attributes
	//	*TestResource_SubResourceAttributes
	Attrs isTestResource_Attrs `protobuf_oneof:"attrs"`
}

func (x *TestResource) Reset() {
	*x = TestResource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_testing_attribute_v1_attribute_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestResource) ProtoMessage() {}

func (x *TestResource) ProtoReflect() protoreflect.Message {
	mi := &file_testing_attribute_v1_attribute_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestResource.ProtoReflect.Descriptor instead.
func (*TestResource) Descriptor() ([]byte, []int) {
	return file_testing_attribute_v1_attribute_proto_rawDescGZIP(), []int{0}
}

func (x *TestResource) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (m *TestResource) GetAttrs() isTestResource_Attrs {
	if m != nil {
		return m.Attrs
	}
	return nil
}

func (x *TestResource) GetAttributes() *structpb.Struct {
	if x, ok := x.GetAttrs().(*TestResource_Attributes); ok {
		return x.Attributes
	}
	return nil
}

func (x *TestResource) GetSubResourceAttributes() *TestSubResourceAttributes {
	if x, ok := x.GetAttrs().(*TestResource_SubResourceAttributes); ok {
		return x.SubResourceAttributes
	}
	return nil
}

type isTestResource_Attrs interface {
	isTestResource_Attrs()
}

type TestResource_Attributes struct {
	Attributes *structpb.Struct `protobuf:"bytes,10,opt,name=attributes,proto3,oneof"`
}

type TestResource_SubResourceAttributes struct {
	SubResourceAttributes *TestSubResourceAttributes `protobuf:"bytes,20,opt,name=sub_resource_attributes,json=subResourceAttributes,proto3,oneof"`
}

func (*TestResource_Attributes) isTestResource_Attrs() {}

func (*TestResource_SubResourceAttributes) isTestResource_Attrs() {}

type TestSubResourceAttributes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *TestSubResourceAttributes) Reset() {
	*x = TestSubResourceAttributes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_testing_attribute_v1_attribute_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestSubResourceAttributes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestSubResourceAttributes) ProtoMessage() {}

func (x *TestSubResourceAttributes) ProtoReflect() protoreflect.Message {
	mi := &file_testing_attribute_v1_attribute_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestSubResourceAttributes.ProtoReflect.Descriptor instead.
func (*TestSubResourceAttributes) Descriptor() ([]byte, []int) {
	return file_testing_attribute_v1_attribute_proto_rawDescGZIP(), []int{1}
}

func (x *TestSubResourceAttributes) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type TestNoAttributes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *TestNoAttributes) Reset() {
	*x = TestNoAttributes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_testing_attribute_v1_attribute_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestNoAttributes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestNoAttributes) ProtoMessage() {}

func (x *TestNoAttributes) ProtoReflect() protoreflect.Message {
	mi := &file_testing_attribute_v1_attribute_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestNoAttributes.ProtoReflect.Descriptor instead.
func (*TestNoAttributes) Descriptor() ([]byte, []int) {
	return file_testing_attribute_v1_attribute_proto_rawDescGZIP(), []int{2}
}

func (x *TestNoAttributes) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

var File_testing_attribute_v1_attribute_proto protoreflect.FileDescriptor

var file_testing_attribute_v1_attribute_proto_rawDesc = []byte{
	0x0a, 0x24, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2f, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62,
	0x75, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e,
	0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74,
	0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2a, 0x63, 0x6f, 0x6e, 0x74,
	0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x6f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xec, 0x01, 0x0a, 0x0c, 0x54, 0x65, 0x73, 0x74, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x46, 0x0a, 0x0a, 0x61, 0x74, 0x74, 0x72, 0x69,
	0x62, 0x75, 0x74, 0x65, 0x73, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74,
	0x72, 0x75, 0x63, 0x74, 0x42, 0x0b, 0x9a, 0xe3, 0x29, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c,
	0x74, 0x48, 0x00, 0x52, 0x0a, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x12,
	0x7b, 0x0a, 0x17, 0x73, 0x75, 0x62, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f,
	0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x18, 0x14, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x2f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x61, 0x74, 0x74, 0x72, 0x69,
	0x62, 0x75, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x53, 0x75, 0x62, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
	0x73, 0x42, 0x10, 0x9a, 0xe3, 0x29, 0x0c, 0x73, 0x75, 0x62, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x48, 0x00, 0x52, 0x15, 0x73, 0x75, 0x62, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x42, 0x07, 0x0a, 0x05,
	0x61, 0x74, 0x74, 0x72, 0x73, 0x22, 0x2f, 0x0a, 0x19, 0x54, 0x65, 0x73, 0x74, 0x53, 0x75, 0x62,
	0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
	0x65, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x22, 0x0a, 0x10, 0x54, 0x65, 0x73, 0x74, 0x4e, 0x6f,
	0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x42, 0x48, 0x5a, 0x46, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x68, 0x61, 0x73, 0x68, 0x69, 0x63, 0x6f,
	0x72, 0x70, 0x2f, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x61, 0x72, 0x79, 0x2f, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67,
	0x2f, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x3b, 0x61, 0x74, 0x74, 0x72, 0x69,
	0x62, 0x75, 0x74, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_testing_attribute_v1_attribute_proto_rawDescOnce sync.Once
	file_testing_attribute_v1_attribute_proto_rawDescData = file_testing_attribute_v1_attribute_proto_rawDesc
)

func file_testing_attribute_v1_attribute_proto_rawDescGZIP() []byte {
	file_testing_attribute_v1_attribute_proto_rawDescOnce.Do(func() {
		file_testing_attribute_v1_attribute_proto_rawDescData = protoimpl.X.CompressGZIP(file_testing_attribute_v1_attribute_proto_rawDescData)
	})
	return file_testing_attribute_v1_attribute_proto_rawDescData
}

var file_testing_attribute_v1_attribute_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_testing_attribute_v1_attribute_proto_goTypes = []interface{}{
	(*TestResource)(nil),              // 0: testing.attribute.v1.TestResource
	(*TestSubResourceAttributes)(nil), // 1: testing.attribute.v1.TestSubResourceAttributes
	(*TestNoAttributes)(nil),          // 2: testing.attribute.v1.TestNoAttributes
	(*structpb.Struct)(nil),           // 3: google.protobuf.Struct
}
var file_testing_attribute_v1_attribute_proto_depIdxs = []int32{
	3, // 0: testing.attribute.v1.TestResource.attributes:type_name -> google.protobuf.Struct
	1, // 1: testing.attribute.v1.TestResource.sub_resource_attributes:type_name -> testing.attribute.v1.TestSubResourceAttributes
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_testing_attribute_v1_attribute_proto_init() }
func file_testing_attribute_v1_attribute_proto_init() {
	if File_testing_attribute_v1_attribute_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_testing_attribute_v1_attribute_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestResource); i {
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
		file_testing_attribute_v1_attribute_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestSubResourceAttributes); i {
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
		file_testing_attribute_v1_attribute_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestNoAttributes); i {
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
	file_testing_attribute_v1_attribute_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*TestResource_Attributes)(nil),
		(*TestResource_SubResourceAttributes)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_testing_attribute_v1_attribute_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_testing_attribute_v1_attribute_proto_goTypes,
		DependencyIndexes: file_testing_attribute_v1_attribute_proto_depIdxs,
		MessageInfos:      file_testing_attribute_v1_attribute_proto_msgTypes,
	}.Build()
	File_testing_attribute_v1_attribute_proto = out.File
	file_testing_attribute_v1_attribute_proto_rawDesc = nil
	file_testing_attribute_v1_attribute_proto_goTypes = nil
	file_testing_attribute_v1_attribute_proto_depIdxs = nil
}
