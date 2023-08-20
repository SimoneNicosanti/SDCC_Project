from enum import Enum
import json


class Method(Enum) :
    GET = 0,
    PUT = 1,
    DEL = 2


class Message :

    def __init__(self, method : Method, fileName : str) -> None:
        self.method : Method = method
        self.fileName : str = fileName


class MessageEncoder(json.JSONEncoder):
    def default(self, obj : Message):
        return {"Method" : obj.method.name , "FileName" : obj.fileName}