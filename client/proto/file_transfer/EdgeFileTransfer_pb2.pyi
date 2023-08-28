from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class ErrorCodes(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    OK: _ClassVar[ErrorCodes]
    CHUNK_ERROR: _ClassVar[ErrorCodes]
    INVALID_TICKET: _ClassVar[ErrorCodes]
    FILE_CREATE_ERROR: _ClassVar[ErrorCodes]
    FILE_NOT_FOUND_ERROR: _ClassVar[ErrorCodes]
    FILE_WRITE_ERROR: _ClassVar[ErrorCodes]
    FILE_READ_ERROR: _ClassVar[ErrorCodes]
    STREAM_CLOSE_ERROR: _ClassVar[ErrorCodes]
OK: ErrorCodes
CHUNK_ERROR: ErrorCodes
INVALID_TICKET: ErrorCodes
FILE_CREATE_ERROR: ErrorCodes
FILE_NOT_FOUND_ERROR: ErrorCodes
FILE_WRITE_ERROR: ErrorCodes
FILE_READ_ERROR: ErrorCodes
STREAM_CLOSE_ERROR: ErrorCodes

class EdgeFileDownloadRequest(_message.Message):
    __slots__ = ["file_name"]
    FILE_NAME_FIELD_NUMBER: _ClassVar[int]
    file_name: str
    def __init__(self, file_name: _Optional[str] = ...) -> None: ...

class EdgeFileChunk(_message.Message):
    __slots__ = ["chunk"]
    CHUNK_FIELD_NUMBER: _ClassVar[int]
    chunk: bytes
    def __init__(self, chunk: _Optional[bytes] = ...) -> None: ...
