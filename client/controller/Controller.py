import concurrent.futures as futures
from engineering import RabbitSingleton
from engineering.Message import Message, Method, MessageEncoder
from proto.file_transfer.File_pb2 import *
from proto.file_transfer.File_pb2_grpc import *
from utils import Utils
import json, pika, grpc, socket, secrets, sys, threading

MAXINT32 = 2147483647
random_request_id = secrets.randbelow(MAXINT32)
sem : threading.Semaphore = threading.Semaphore(0)

def sendRequestForFile(requestType, fileName):
    global random_request_id
    channel = RabbitSingleton.getRabbitChannel()
    channel.queue_declare(queue = 'storage_queue', durable=True)

    print(socket.gethostbyname(socket.gethostname())+':50051')
    random_request_id = secrets.randbelow(MAXINT32)
    message : Message = Message(random_request_id, requestType, fileName, socket.gethostbyname(socket.gethostname())+':50051')
    channel.basic_publish(
        exchange = '', 
        routing_key = 'storage_queue', 
        body = json.dumps(obj = message, cls = MessageEncoder),
        properties = pika.BasicProperties(delivery_mode = pika.spec.PERSISTENT_DELIVERY_MODE)
    )
    print(f"[*] {requestType} request sent for file '{fileName}' [REQ_ID '{random_request_id}']")

    sem.acquire()
    #TODO wait for response

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
    add_FileServiceServicer_to_server(FileService(), server)
    server.add_insecure_port('[::]:50051')
    server.start()

def getFile(fileName : str) -> None :
    sendRequestForFile(Method.GET, fileName)


def putFile(fileName : str) -> None :
    sendRequestForFile(Method.PUT, fileName)


def deleteFile(fileName : str) -> None:
    sendRequestForFile(Method.DEL, fileName)


def login(username : str, passwd : str, email : str) -> bool :
    if username == "sae" and passwd == "admin" and email == "admin@sae.com":
        serve()
        return True


class FileService(FileServiceServicer):
    global random_request_id

    def Upload(self, request : FileUploadRequest, context):
        #Security check
        if request.request_id != random_request_id:
            return

        with open(request.filename, "rb") as file :
            chunkSize = int(Utils.readProperties("conf.properties", "CHUNK_SIZE"))
            chunk = file.read(chunkSize)
            while chunk:
                fileChunk : FileChunk = FileChunk(file_name = request.file_name, file_chunk = chunk)
                yield fileChunk
                chunk = file.read(chunkSize)

            sem.release()

    def Download(self, request_iterator, context) -> Response :
        
        def secure_write(request_id, file, chunk):
            #Security check
            if request_id != random_request_id:
                return False
            file.write(chunk.file_chunk)
        #-----------------------------------------

        chunk : FileChunk = next(request_iterator)
        filename = chunk.file_name
        print("RECEIVED DOWNLOAD")

        with open(filename, "w") as file:
            if not secure_write(file, chunk.file_chunk):
                    return Response(request_id = random_request_id, success = False)
            for chunk in request_iterator:
                chunk : FileChunk
                
                if not secure_write(file, chunk.file_chunk):
                    return Response(request_id = random_request_id, success = False)
                          
        sem.release()
        return Response(request_id = random_request_id, success = True)
    
    
        
        