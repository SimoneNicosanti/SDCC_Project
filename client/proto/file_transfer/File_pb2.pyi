from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class FileUploadRequest(_message.Message):
    __slots__ = ["file_name"]
    FILE_NAME_FIELD_NUMBER: _ClassVar[int]
    file_name: str
    def __init__(self, file_name: _Optional[str] = ...) -> None: ...

class FileChunk(_message.Message):
    __slots__ = ["file_name", "file_chunk"]
    FILE_NAME_FIELD_NUMBER: _ClassVar[int]
    FILE_CHUNK_FIELD_NUMBER: _ClassVar[int]
    file_name: str
    file_chunk: bytes
    def __init__(self, file_name: _Optional[str] = ..., file_chunk: _Optional[bytes] = ...) -> None: ...

class Response(_message.Message):
    __slots__ = ["success"]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    success: bool
    def __init__(self, success: bool = ...) -> None: ...

class FileDownloadResponse(_message.Message):
    __slots__ = ["file_name", "file_size"]
    FILE_NAME_FIELD_NUMBER: _ClassVar[int]
    FILE_SIZE_FIELD_NUMBER: _ClassVar[int]
    file_name: str
    file_size: int
    def __init__(self, file_name: _Optional[str] = ..., file_size: _Optional[int] = ...) -> None: ...
