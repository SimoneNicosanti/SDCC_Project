from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class ErrorCodes(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    OK: _ClassVar[ErrorCodes]
    S3_ERROR: _ClassVar[ErrorCodes]
    CHUNK_ERROR: _ClassVar[ErrorCodes]
    INVALID_TICKET: _ClassVar[ErrorCodes]
    FILE_CREATE_ERROR: _ClassVar[ErrorCodes]
    FILE_NOT_FOUND_ERROR: _ClassVar[ErrorCodes]
    FILE_WRITE_ERROR: _ClassVar[ErrorCodes]
    FILE_READ_ERROR: _ClassVar[ErrorCodes]
    STREAM_CLOSE_ERROR: _ClassVar[ErrorCodes]
OK: ErrorCodes
S3_ERROR: ErrorCodes
CHUNK_ERROR: ErrorCodes
INVALID_TICKET: ErrorCodes
FILE_CREATE_ERROR: ErrorCodes
FILE_NOT_FOUND_ERROR: ErrorCodes
FILE_WRITE_ERROR: ErrorCodes
FILE_READ_ERROR: ErrorCodes
STREAM_CLOSE_ERROR: ErrorCodes

class FileDownloadRequest(_message.Message):
    __slots__ = ["ticket_id", "file_name"]
    TICKET_ID_FIELD_NUMBER: _ClassVar[int]
    FILE_NAME_FIELD_NUMBER: _ClassVar[int]
    ticket_id: str
    file_name: str
    def __init__(self, ticket_id: _Optional[str] = ..., file_name: _Optional[str] = ...) -> None: ...

class FileChunk(_message.Message):
    __slots__ = ["chunk"]
    CHUNK_FIELD_NUMBER: _ClassVar[int]
    chunk: bytes
    def __init__(self, chunk: _Optional[bytes] = ...) -> None: ...

class Response(_message.Message):
    __slots__ = ["ticket_id", "success"]
    TICKET_ID_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    ticket_id: str
    success: bool
    def __init__(self, ticket_id: _Optional[str] = ..., success: bool = ...) -> None: ...
