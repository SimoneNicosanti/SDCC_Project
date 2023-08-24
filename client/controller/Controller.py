import concurrent.futures as futures
from engineering import RabbitSingleton
from engineering.Message import Message, Method, MessageEncoder
from proto.file_transfer.File_pb2 import *
from proto.file_transfer.File_pb2_grpc import *
from utils import Utils
import json, pika, grpc, socket, secrets, sys, threading, ssl, sched, time, os

MAXINT32 = 2147483647
random_request_id : list = []

def sendRequestForFile(requestType, fileName):
    global random_request_id
    channel = RabbitSingleton.getRabbitChannel()
    channel.queue_declare(queue = 'storage_queue', durable=True)

    print(socket.gethostbyname(socket.gethostname())+':50051')
    current_request_id = secrets.randbelow(MAXINT32)
    random_request_id.append(current_request_id)
    message : Message = Message(random_request_id, requestType, fileName, socket.gethostbyname(socket.gethostname())+':50051')
    channel.basic_publish(
        exchange = '', 
        routing_key = 'storage_queue', 
        body = json.dumps(obj = message, cls = MessageEncoder),
        properties = pika.BasicProperties(delivery_mode = pika.spec.PERSISTENT_DELIVERY_MODE)
    )

    timer = threading.Timer(int(os.environ.get("TTL")), checkSatisfiedRequest(request_id = current_request_id, filename = fileName))
    timer.start()
    

    print(f"[*] {requestType} request sent for file '{fileName}' [REQ_ID '{current_request_id}']")

def checkSatisfiedRequest(request_id : int, filename : str) -> None:
    if request_id in random_request_id:
        print("La richiesta per il file '{filename}' non è stata soddisfatta. Prova ad effettuare una nuova richiesta.")
        random_request_id.remove(request_id)

def serve():
    #TODO TLS CONFIGURATION =======================================================
    # Load your server's certificate and private key
    #server_cert = open("/src/tls/server-cert.pem", "rb").read()
    #server_key = open("/src/tls/server-key.pem", "rb").read()
    # Create server credentials using the loaded certificate and key
    #server_credentials = grpc.ssl_server_credentials(((server_key, server_cert),))
    #==============================================================================

    # Create a gRPC server with secure credentials
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
    add_FileServiceServicer_to_server(FileService(), server)
    server.add_insecure_port('[::]:50051')
    #TODO server.add_secure_port('[::]:50051', server_credentials)
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
        if request.request_id not in random_request_id:
            print(f"[*ERROR*] - Security check failed --> {request.request_id} is not in the expected request list")
            return

        try:
            with open("/files/" + request.file_name, "rb") as file :
                chunkSize = int(Utils.readProperties("conf.properties", "CHUNK_SIZE"))
                chunk = file.read(chunkSize)
                while chunk:
                    fileChunk : FileChunk = FileChunk(request_id = request.request_id, file_name = request.file_name, chunk = chunk)
                    yield fileChunk
                    chunk = file.read(chunkSize)
        except IOError as e:
            print("[*ERROR*] - Couldn't open the file:", str(e))
        finally:
            random_request_id.remove(request.request_id)

        random_request_id.remove(request.request_id)


    def Download(self, request_iterator, context) -> Response :
        
        try:
            response : Response = self.downloadFile(request_iterator)
        except IOError as e:
            print("[*ERROR*] - downloadFile(request_iterator) failed", str(e))
        
        return response

    def downloadFile(request_iterator) -> Response:
        def secure_write(request_id, file, chunk, current_request_id = None) -> bool:
            #Security check
            if current_request_id == None:
                if request_id not in random_request_id:
                    print(f"[*ERROR*] - Security check failed --> {request_id} is not in the expected request list")
                    return False
            else:
                if request_id != current_request_id:
                    print(f"[*ERROR*] - Security check failed --> one of the chunk has a different request_id: {request_id}(current_chunk)!={current_request_id}(other_chunk)")
                    return False
            file.write(chunk)
            return True
        #-----------------------------------------

        chunk : FileChunk = next(request_iterator)
        filename = chunk.file_name
        current_request_id = chunk.request_id
        try:
            with open("/files/" + filename, "wb") as file:
                if not secure_write(chunk.request_id, file, chunk.chunk):
                    return Response(request_id = chunk.request_id, success = False)
                for chunk in request_iterator:
                    chunk : FileChunk
                    
                    if not secure_write(chunk.request_id, file, chunk.chunk):
                        return Response(request_id = chunk.request_id, success = False, current_request_id = current_request_id)
        except IOError as e:
            print("[*ERROR*] - Couldn't open the file:", str(e))

        random_request_id.remove(current_request_id)
        return Response(request_id = chunk.request_id, success = True)
    
    
        
        