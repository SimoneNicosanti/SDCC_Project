from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class FileDownloadRequest(_message.Message):
    __slots__ = ["request_id", "file_name"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    FILE_NAME_FIELD_NUMBER: _ClassVar[int]
    request_id: int
    file_name: str
    def __init__(self, request_id: _Optional[int] = ..., file_name: _Optional[str] = ...) -> None: ...

class FileChunk(_message.Message):
    __slots__ = ["request_id", "file_name", "chunk"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    FILE_NAME_FIELD_NUMBER: _ClassVar[int]
    CHUNK_FIELD_NUMBER: _ClassVar[int]
    request_id: int
    file_name: str
    chunk: bytes
    def __init__(self, request_id: _Optional[int] = ..., file_name: _Optional[str] = ..., chunk: _Optional[bytes] = ...) -> None: ...

class Response(_message.Message):
    __slots__ = ["request_id", "success"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    request_id: int
    success: bool
    def __init__(self, request_id: _Optional[int] = ..., success: bool = ...) -> None: ...
