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
    INVALID_METADATA: _ClassVar[ErrorCodes]
    FILE_CREATE_ERROR: _ClassVar[ErrorCodes]
    FILE_NOT_FOUND_ERROR: _ClassVar[ErrorCodes]
    FILE_WRITE_ERROR: _ClassVar[ErrorCodes]
    FILE_READ_ERROR: _ClassVar[ErrorCodes]
    STREAM_CLOSE_ERROR: _ClassVar[ErrorCodes]
    REQUEST_FAILED: _ClassVar[ErrorCodes]
OK: ErrorCodes
S3_ERROR: ErrorCodes
CHUNK_ERROR: ErrorCodes
INVALID_METADATA: ErrorCodes
FILE_CREATE_ERROR: ErrorCodes
FILE_NOT_FOUND_ERROR: ErrorCodes
FILE_WRITE_ERROR: ErrorCodes
FILE_READ_ERROR: ErrorCodes
STREAM_CLOSE_ERROR: ErrorCodes
REQUEST_FAILED: ErrorCodes

class FileDownloadRequest(_message.Message):
    __slots__ = ["fileName", "requestId"]
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    REQUESTID_FIELD_NUMBER: _ClassVar[int]
    fileName: str
    requestId: str
    def __init__(self, fileName: _Optional[str] = ..., requestId: _Optional[str] = ...) -> None: ...

class FileChunk(_message.Message):
    __slots__ = ["chunk", "seqNum"]
    CHUNK_FIELD_NUMBER: _ClassVar[int]
    SEQNUM_FIELD_NUMBER: _ClassVar[int]
    chunk: bytes
    seqNum: int
    def __init__(self, chunk: _Optional[bytes] = ..., seqNum: _Optional[int] = ...) -> None: ...

class FileResponse(_message.Message):
    __slots__ = ["requestId", "success"]
    REQUESTID_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    requestId: str
    success: bool
    def __init__(self, requestId: _Optional[str] = ..., success: bool = ...) -> None: ...

class FileDeleteRequest(_message.Message):
    __slots__ = ["fileName", "requestId"]
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    REQUESTID_FIELD_NUMBER: _ClassVar[int]
    fileName: str
    requestId: str
    def __init__(self, fileName: _Optional[str] = ..., requestId: _Optional[str] = ...) -> None: ...
