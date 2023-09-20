from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class LoginResponse(_message.Message):
    __slots__ = ["logged"]
    LOGGED_FIELD_NUMBER: _ClassVar[int]
    logged: bool
    def __init__(self, logged: bool = ...) -> None: ...

class User(_message.Message):
    __slots__ = ["username", "passwd"]
    USERNAME_FIELD_NUMBER: _ClassVar[int]
    PASSWD_FIELD_NUMBER: _ClassVar[int]
    username: str
    passwd: str
    def __init__(self, username: _Optional[str] = ..., passwd: _Optional[str] = ...) -> None: ...

class BalancerResponse(_message.Message):
    __slots__ = ["success", "edgeIpAddr", "requestId"]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    EDGEIPADDR_FIELD_NUMBER: _ClassVar[int]
    REQUESTID_FIELD_NUMBER: _ClassVar[int]
    success: bool
    edgeIpAddr: str
    requestId: str
    def __init__(self, success: bool = ..., edgeIpAddr: _Optional[str] = ..., requestId: _Optional[str] = ...) -> None: ...
