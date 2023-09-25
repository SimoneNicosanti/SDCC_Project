from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class ErrorCodesLoadBalancer(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    OK: _ClassVar[ErrorCodesLoadBalancer]
    NO_SERVER_AVAILABLE: _ClassVar[ErrorCodesLoadBalancer]
OK: ErrorCodesLoadBalancer
NO_SERVER_AVAILABLE: ErrorCodesLoadBalancer

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
    __slots__ = ["edgeIpAddr", "requestId"]
    EDGEIPADDR_FIELD_NUMBER: _ClassVar[int]
    REQUESTID_FIELD_NUMBER: _ClassVar[int]
    edgeIpAddr: str
    requestId: str
    def __init__(self, edgeIpAddr: _Optional[str] = ..., requestId: _Optional[str] = ...) -> None: ...
