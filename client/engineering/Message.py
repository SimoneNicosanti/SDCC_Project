from enum import Enum
import json


class Method(Enum) :
    GET = 0,
    PUT = 1,
    DEL = 2


class Message :

    def __init__(self, request_id : int, method : Method, fileName : str, ip_addr : str) -> None:
        self.request_id : int = request_id
        self.method : Method = method
        self.fileName : str = fileName
        self.ip_addr : str = ip_addr


class MessageEncoder(json.JSONEncoder):
    def default(self, obj : Message):
        return {"RequestId" : obj.request_id, "Method" : obj.method.name , "FileName" : obj.fileName, "IpAddr" : obj.ip_addr}