from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class UserInfo(_message.Message):
    __slots__ = ["username", "passwd", "email"]
    USERNAME_FIELD_NUMBER: _ClassVar[int]
    PASSWD_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    username: str
    passwd: str
    email: str
    def __init__(self, username: _Optional[str] = ..., passwd: _Optional[str] = ..., email: _Optional[str] = ...) -> None: ...

class Response(_message.Message):
    __slots__ = ["response"]
    class ResponseType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        SUCCESS: _ClassVar[Response.ResponseType]
        ERROR: _ClassVar[Response.ResponseType]
    SUCCESS: Response.ResponseType
    ERROR: Response.ResponseType
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    response: Response.ResponseType
    def __init__(self, response: _Optional[_Union[Response.ResponseType, str]] = ...) -> None: ...
