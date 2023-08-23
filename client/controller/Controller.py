from engineering import RabbitSingleton
from engineering.Message import Message, Method, MessageEncoder
from proto.file_transfer import File_pb2, File_pb2_grpc
import json, pika, grpc
from jproperties import Properties


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


def startServer() :
    pass

def getFile(fileName : str) -> None :
    sendRequestForFile(Method.GET, "file_di_prova.txt")


def putFile(fileName : str) -> None :
    sendRequestForFile(Method.PUT, "file_di_prova.txt")


def deleteFile(fileName : str) -> None:
    sendRequestForFile(Method.DEL, "file_di_prova.txt")


def login(username : str, passwd : str, email : str) -> bool :
    return username == "sae" and passwd == "admin" and email == "admin@sae.com"


class FileService(File_pb2_grpc.FileServiceServicer):

    def Upload(self, request : File_pb2.FileUploadRequest, context):
        fileChunks : list = []
        file = open(request.filename, "rb")
        chunk = file.read(CHUNK_SIZE)
        for elem in fileChunks :
            fileChunk : File_pb2 = File_pb2.FileChunk(file_chunk = elem)
            yield fileChunk

    def Download(self, request_iterator, context):
        """rpc Delete(FileDeleteRequest) returns (FileDeleteResponse)
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')