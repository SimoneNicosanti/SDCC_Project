from engineering import RabbitSingleton
from engineering.Message import Message, Method, MessageEncoder
import json, pika

def sendRequestForFile(requestType, fileName):
    channel = RabbitSingleton.getRabbitChannel()
    channel.queue_declare(queue = 'storage_queue', durable=True)

    message : Message = Message(requestType, fileName)
    channel.basic_publish(
        exchange = '', 
        routing_key = 'storage_queue', 
        body = json.dumps(obj = message, cls = MessageEncoder),
        properties = pika.BasicProperties(delivery_mode = pika.spec.PERSISTENT_DELIVERY_MODE)
    )
    print(f"[*] {requestType} request sent for file '{fileName}'")

def getFile(fileName : str) -> None :
    sendRequestForFile(Method.GET, "file_di_prova.txt")


def putFile(fileName : str) -> None :
    sendRequestForFile(Method.PUT, "file_di_prova.txt")


def deleteFile(fileName : str) -> None:
    sendRequestForFile(Method.DEL, "file_di_prova.txt")


def login(username : str, passwd : str, email : str) -> bool :
    return username == "sae" and passwd == "admin" and email == "admin@sae.com"