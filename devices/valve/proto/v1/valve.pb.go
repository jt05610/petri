// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.23.4
// source: v1/valve.proto

package valve

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Command int32

const (
	Command_OpenA  Command = 0
	Command_CloseA Command = 1
)

// Enum value maps for Command.
var (
	Command_name = map[int32]string{
		0: "OpenA",
		1: "CloseA",
	}
	Command_value = map[string]int32{
		"OpenA":  0,
		"CloseA": 1,
	}
)

func (x Command) Enum() *Command {
	p := new(Command)
	*p = x
	return p
}

func (x Command) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Command) Descriptor() protoreflect.EnumDescriptor {
	return file_v1_valve_proto_enumTypes[0].Descriptor()
}

func (Command) Type() protoreflect.EnumType {
	return &file_v1_valve_proto_enumTypes[0]
}

func (x Command) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Command.Descriptor instead.
func (Command) EnumDescriptor() ([]byte, []int) {
	return file_v1_valve_proto_rawDescGZIP(), []int{0}
}

type Event int32

const (
	Event_OpenedA Event = 0
	Event_ClosedA Event = 1
)

// Enum value maps for Event.
var (
	Event_name = map[int32]string{
		0: "OpenedA",
		1: "ClosedA",
	}
	Event_value = map[string]int32{
		"OpenedA": 0,
		"ClosedA": 1,
	}
)

func (x Event) Enum() *Event {
	p := new(Event)
	*p = x
	return p
}

func (x Event) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Event) Descriptor() protoreflect.EnumDescriptor {
	return file_v1_valve_proto_enumTypes[1].Descriptor()
}

func (Event) Type() protoreflect.EnumType {
	return &file_v1_valve_proto_enumTypes[1]
}

func (x Event) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Event.Descriptor instead.
func (Event) EnumDescriptor() ([]byte, []int) {
	return file_v1_valve_proto_rawDescGZIP(), []int{1}
}

type State struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	A int32 `protobuf:"varint,1,opt,name=a,proto3" json:"a,omitempty"`
	B int32 `protobuf:"varint,2,opt,name=b,proto3" json:"b,omitempty"`
}

func (x *State) Reset() {
	*x = State{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_valve_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *State) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*State) ProtoMessage() {}

func (x *State) ProtoReflect() protoreflect.Message {
	mi := &file_v1_valve_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use State.ProtoReflect.Descriptor instead.
func (*State) Descriptor() ([]byte, []int) {
	return file_v1_valve_proto_rawDescGZIP(), []int{0}
}

func (x *State) GetA() int32 {
	if x != nil {
		return x.A
	}
	return 0
}

func (x *State) GetB() int32 {
	if x != nil {
		return x.B
	}
	return 0
}

type ValveStateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State *State `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
}

func (x *ValveStateResponse) Reset() {
	*x = ValveStateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_valve_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValveStateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValveStateResponse) ProtoMessage() {}

func (x *ValveStateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_valve_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValveStateResponse.ProtoReflect.Descriptor instead.
func (*ValveStateResponse) Descriptor() ([]byte, []int) {
	return file_v1_valve_proto_rawDescGZIP(), []int{1}
}

func (x *ValveStateResponse) GetState() *State {
	if x != nil {
		return x.State
	}
	return nil
}

type ValveEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Event Event      `protobuf:"varint,1,opt,name=event,proto3,enum=valve.Event" json:"event,omitempty"`
	Data  *anypb.Any `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *ValveEvent) Reset() {
	*x = ValveEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_valve_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValveEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValveEvent) ProtoMessage() {}

func (x *ValveEvent) ProtoReflect() protoreflect.Message {
	mi := &file_v1_valve_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValveEvent.ProtoReflect.Descriptor instead.
func (*ValveEvent) Descriptor() ([]byte, []int) {
	return file_v1_valve_proto_rawDescGZIP(), []int{2}
}

func (x *ValveEvent) GetEvent() Event {
	if x != nil {
		return x.Event
	}
	return Event_OpenedA
}

func (x *ValveEvent) GetData() *anypb.Any {
	if x != nil {
		return x.Data
	}
	return nil
}

type ValveCommand struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Command Command    `protobuf:"varint,1,opt,name=command,proto3,enum=valve.Command" json:"command,omitempty"`
	Data    *anypb.Any `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *ValveCommand) Reset() {
	*x = ValveCommand{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_valve_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValveCommand) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValveCommand) ProtoMessage() {}

func (x *ValveCommand) ProtoReflect() protoreflect.Message {
	mi := &file_v1_valve_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValveCommand.ProtoReflect.Descriptor instead.
func (*ValveCommand) Descriptor() ([]byte, []int) {
	return file_v1_valve_proto_rawDescGZIP(), []int{3}
}

func (x *ValveCommand) GetCommand() Command {
	if x != nil {
		return x.Command
	}
	return Command_OpenA
}

func (x *ValveCommand) GetData() *anypb.Any {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_v1_valve_proto protoreflect.FileDescriptor

var file_v1_valve_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x6c, 0x76, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x05, 0x76, 0x61, 0x6c, 0x76, 0x65, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x23, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0c, 0x0a, 0x01, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x01, 0x61, 0x12, 0x0c, 0x0a, 0x01, 0x62, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x01, 0x62, 0x22, 0x38, 0x0a, 0x12, 0x56, 0x61, 0x6c, 0x76, 0x65, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x22, 0x0a, 0x05, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x76, 0x61, 0x6c, 0x76,
	0x65, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x22, 0x5a,
	0x0a, 0x0a, 0x56, 0x61, 0x6c, 0x76, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x22, 0x0a, 0x05,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x76, 0x61,
	0x6c, 0x76, 0x65, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x12, 0x28, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x41, 0x6e, 0x79, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x62, 0x0a, 0x0c, 0x56, 0x61,
	0x6c, 0x76, 0x65, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x28, 0x0a, 0x07, 0x63, 0x6f,
	0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x76, 0x61,
	0x6c, 0x76, 0x65, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x07, 0x63, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x12, 0x28, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x2a, 0x20,
	0x0a, 0x07, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x09, 0x0a, 0x05, 0x4f, 0x70, 0x65,
	0x6e, 0x41, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x41, 0x10, 0x01,
	0x2a, 0x21, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x0b, 0x0a, 0x07, 0x4f, 0x70, 0x65,
	0x6e, 0x65, 0x64, 0x41, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x64,
	0x41, 0x10, 0x01, 0x32, 0x84, 0x01, 0x0a, 0x0c, 0x56, 0x61, 0x6c, 0x76, 0x65, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x3d, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x19, 0x2e, 0x76, 0x61, 0x6c, 0x76, 0x65,
	0x2e, 0x56, 0x61, 0x6c, 0x76, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x35, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x12, 0x13,
	0x2e, 0x76, 0x61, 0x6c, 0x76, 0x65, 0x2e, 0x56, 0x61, 0x6c, 0x76, 0x65, 0x43, 0x6f, 0x6d, 0x6d,
	0x61, 0x6e, 0x64, 0x1a, 0x11, 0x2e, 0x76, 0x61, 0x6c, 0x76, 0x65, 0x2e, 0x56, 0x61, 0x6c, 0x76,
	0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x28, 0x01, 0x30, 0x01, 0x42, 0x37, 0x5a, 0x35, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x74, 0x30, 0x35, 0x36, 0x31, 0x30,
	0x2f, 0x70, 0x65, 0x74, 0x72, 0x69, 0x2f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2f, 0x76,
	0x61, 0x6c, 0x76, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61,
	0x6c, 0x76, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_v1_valve_proto_rawDescOnce sync.Once
	file_v1_valve_proto_rawDescData = file_v1_valve_proto_rawDesc
)

func file_v1_valve_proto_rawDescGZIP() []byte {
	file_v1_valve_proto_rawDescOnce.Do(func() {
		file_v1_valve_proto_rawDescData = protoimpl.X.CompressGZIP(file_v1_valve_proto_rawDescData)
	})
	return file_v1_valve_proto_rawDescData
}

var file_v1_valve_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_v1_valve_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_v1_valve_proto_goTypes = []interface{}{
	(Command)(0),               // 0: valve.Command
	(Event)(0),                 // 1: valve.Event
	(*State)(nil),              // 2: valve.State
	(*ValveStateResponse)(nil), // 3: valve.ValveStateResponse
	(*ValveEvent)(nil),         // 4: valve.ValveEvent
	(*ValveCommand)(nil),       // 5: valve.ValveCommand
	(*anypb.Any)(nil),          // 6: google.protobuf.Any
	(*emptypb.Empty)(nil),      // 7: google.protobuf.Empty
}
var file_v1_valve_proto_depIdxs = []int32{
	2, // 0: valve.ValveStateResponse.state:type_name -> valve.State
	1, // 1: valve.ValveEvent.event:type_name -> valve.Event
	6, // 2: valve.ValveEvent.data:type_name -> google.protobuf.Any
	0, // 3: valve.ValveCommand.command:type_name -> valve.Command
	6, // 4: valve.ValveCommand.data:type_name -> google.protobuf.Any
	7, // 5: valve.ValveService.GetState:input_type -> google.protobuf.Empty
	5, // 6: valve.ValveService.Control:input_type -> valve.ValveCommand
	3, // 7: valve.ValveService.GetState:output_type -> valve.ValveStateResponse
	4, // 8: valve.ValveService.Control:output_type -> valve.ValveEvent
	7, // [7:9] is the sub-list for method output_type
	5, // [5:7] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_v1_valve_proto_init() }
func file_v1_valve_proto_init() {
	if File_v1_valve_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_v1_valve_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*State); i {
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
		file_v1_valve_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValveStateResponse); i {
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
		file_v1_valve_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValveEvent); i {
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
		file_v1_valve_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValveCommand); i {
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
			RawDescriptor: file_v1_valve_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_v1_valve_proto_goTypes,
		DependencyIndexes: file_v1_valve_proto_depIdxs,
		EnumInfos:         file_v1_valve_proto_enumTypes,
		MessageInfos:      file_v1_valve_proto_msgTypes,
	}.Build()
	File_v1_valve_proto = out.File
	file_v1_valve_proto_rawDesc = nil
	file_v1_valve_proto_goTypes = nil
	file_v1_valve_proto_depIdxs = nil
}
