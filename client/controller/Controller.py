import concurrent.futures as futures
from engineering import RabbitSingleton
from engineering.Message import Message, Method, MessageEncoder
from proto.file_transfer.File_pb2 import *
from proto.file_transfer.File_pb2_grpc import *
from utils import Utils
import json, pika, grpc, socket, secrets, threading, ssl, os

MAXINT32 = 2147483647
random_request_id : list = []

def sendRequestForFile(requestType : Method, fileName : str) -> None:

    # Otteniamo l'indirizzo del peer da contattare
    new_id = secrets.randbelow(MAXINT32)
    peer_addr = getPeer(requestType = requestType, current_request_id = new_id)

    channel = grpc.insecure_channel(IP DA CONTATTARE)
    stub = FileServiceStub(channel)

    execAction(requestId = new_id, requestType = requestType, filename = fileName, stub = stub)

    timer = threading.Timer(int(os.environ.get("TTL")), checkSatisfiedRequest(request_id = new_id, filename = fileName))
    timer.start()
    

    print(f"[*] {requestType} request sent for file '{fileName}' [REQ_ID '{new_id}']")

def getPeer(requestType : Method, current_request_id : int) -> str:
    global random_request_id

    queue_name = os.environ.get("QUEUE_NAME")
    channel = RabbitSingleton.getRabbitChannel()
    channel.queue_declare(queue = queue_name, durable=True)

    print(socket.gethostbyname(socket.gethostname())+':50051')
    
    random_request_id.append(current_request_id)

    #TODO modificare logica coda: ora il client legge il peer da contattare
    message : Message = Message(random_request_id, requestType, fileName, socket.gethostbyname(socket.gethostname())+':50051')
    channel.basic_publish(
        exchange = '', 
        routing_key = queue_name, 
        body = json.dumps(obj = message, cls = MessageEncoder),
        properties = pika.BasicProperties(delivery_mode = pika.spec.PERSISTENT_DELIVERY_MODE)
    )


def execAction(request_id : int, requestType : Method, filename : str, stub : FileServiceStub) -> bool:

    switch : dict = {
        Method.GET : getFile,
        Method.PUT : putFile,
        Method.DEL : deleteFile
    }

    action = switch(requestType)

    result = action(filename, request_id, stub)
    
    if result == True:
        random_request_id.remove(request_id)
    
    return result

def getFile(filename : str, request_id : int, stub : FileServiceStub) -> bool :
    # Otteniamo i chunks dalla chiamata gRPC
    chunks = stub.Download(FileDownloadRequest(request_id=request_id, file_name=filename))

    # Scriviamo i chunks in un file e ritorniamo l'esito della scrittura
    return FileService().writeChunks(chunks)


def putFile(fileName : str, request_id : int, stub : FileServiceStub) -> bool :
    # Dividiamo il file in chunks
    chunks = FileService.getChunks()

    # Effettuiamo la chiamata gRPC 
    response : Response = stub.Upload(chunks)

    # Se la risposta non riguarda la nostra richiesta, lanciamo una eccezione
    if response.request_id != request_id:
        raise Exception()
    
    return response.success


def deleteFile(fileName : str) -> bool:
    #TODO
    pass

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


def login(username : str, passwd : str, email : str) -> bool :
    if username == "sae" and passwd == "admin" and email == "admin@sae.com":
        serve()
        return True


class FileService:
    global random_request_id

    def getChunks(self, request_id : int, filename : str):
        #Security check
        if request_id not in random_request_id:
            print(f"[*ERROR*] - Security check failed --> {request_id} is not in the expected request list")
            return

        try:
            with open("/files/" + filename, "rb") as file :
                chunkSize = int(Utils.readProperties("conf.properties", "CHUNK_SIZE"))
                chunk = file.read(chunkSize)
                while chunk:
                    fileChunk : FileChunk = FileChunk(request_id = request_id, file_name = filename, chunk = chunk)
                    yield fileChunk
                    chunk = file.read(chunkSize)
        except IOError as e:
            print("[*ERROR*] - Couldn't open the file:", str(e))
        

    def writeChunks(self, request_iterator):
        
        try:
            return self.downloadFile(request_iterator)
        except IOError as e:
            print("[*ERROR*] - downloadFile(request_iterator) failed", str(e))
            return False

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
                    return False
                for chunk in request_iterator:
                    chunk : FileChunk
                    
                    if not secure_write(chunk.request_id, file, chunk.chunk):
                        return False
        except IOError as e:
            print("[*ERROR*] - Couldn't open the file:", str(e))

        random_request_id.remove(current_request_id)
        return True
    
    
        
        