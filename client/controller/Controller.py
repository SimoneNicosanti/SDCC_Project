from engineering import RabbitSingleton
from engineering.Message import Message, Method, MessageEncoder
import json

def getFile(fileName : str) -> None :
    channel = RabbitSingleton.getRabbitChannel()
    channel.queue_declare(queue = 'storage_queue')

    message : Message = Message(Method.GET, "file_prova")
    channel.basic_publish(
        exchange = '', 
        routing_key = 'storage_queue', 
        body = json.dumps(obj = message, cls = MessageEncoder)
    )
    print("Sent Message")
    pass


def putFile(fileName : str) -> None :
    pass


def deleteFile(fileName : str) -> None:
    pass


def login(username : str, passwd : str, email : str) -> bool :
    return username == "sae" and passwd == "admin" and email == "admin@sae.com"