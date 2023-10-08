from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class InjectState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    Idle: _ClassVar[InjectState]
    Injecting: _ClassVar[InjectState]
    Injected: _ClassVar[InjectState]
Idle: InjectState
Injecting: InjectState
Injected: InjectState

class InjectRequest(_message.Message):
    __slots__ = ["vial", "air_cushion", "excess_volume", "flush_volume", "injection_volume", "needle_depth"]
    VIAL_FIELD_NUMBER: _ClassVar[int]
    AIR_CUSHION_FIELD_NUMBER: _ClassVar[int]
    EXCESS_VOLUME_FIELD_NUMBER: _ClassVar[int]
    FLUSH_VOLUME_FIELD_NUMBER: _ClassVar[int]
    INJECTION_VOLUME_FIELD_NUMBER: _ClassVar[int]
    NEEDLE_DEPTH_FIELD_NUMBER: _ClassVar[int]
    vial: int
    air_cushion: int
    excess_volume: int
    flush_volume: int
    injection_volume: int
    needle_depth: int
    def __init__(self, vial: _Optional[int] = ..., air_cushion: _Optional[int] = ..., excess_volume: _Optional[int] = ..., flush_volume: _Optional[int] = ..., injection_volume: _Optional[int] = ..., needle_depth: _Optional[int] = ...) -> None: ...

class InjectResponse(_message.Message):
    __slots__ = ["state"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    state: InjectState
    def __init__(self, state: _Optional[_Union[InjectState, str]] = ...) -> None: ...
