from engineering import Debug, MyErrors
from engineering.Method import Method
from proto.file_transfer.FileTransfer_pb2 import *
from proto.file_transfer.FileTransfer_pb2_grpc import *
from proto.load_balancer.LoadBalancer_pb2 import *
from proto.load_balancer.LoadBalancer_pb2_grpc import *
from grpc import StatusCode
from asyncio import Semaphore
import json, grpc, os, io, threading

data = None
userInfo : User = User(username="a",passwd="a")
sem : Semaphore = Semaphore()

def sendRequestForFile(requestType : Method, fileName : str) -> bool:
    if requestType == Method.PUT and not os.path.exists(os.environ.get("FILES_PATH") + fileName):
        raise MyErrors.FileNotFound("Il file da caricare non esiste in locale.")
    # Otteniamo l'indirizzo del peer da contattare
    response: BalancerResponse = getEdgeFromBalancer()
    if not response.success:
        raise MyErrors.NoServerAvailable(e.details())
    # Preparazione chiamata gRPC
    try:
        channel : grpc.Channel = grpc.insecure_channel(response.edgeIpAddr, options=[('grpc.max_receive_message_length', int(os.environ.get("MAX_GRPC_MESSAGE_SIZE")))])
        stub = FileServiceStub(channel)
    except grpc.RpcError as e:
        if e.code() == StatusCode.UNAVAILABLE:
            raise MyErrors.ConnectionFailedException(e.details())
        raise e
    # Esecuzione dell'azione
    result = execAction(requestId = response.requestId, requestType = requestType, fileName = fileName, stub = stub)

    return result

def getEdgeFromBalancer() -> BalancerResponse:
    try:
        channel = grpc.insecure_channel(os.environ.get("LOAD_BALANCER_ADDR"))
        stub = BalancingServiceStub(channel)
        return stub.GetEdge(userInfo)
    except grpc.RpcError as e:
        if e.code() == StatusCode.UNAUTHENTICATED:
            raise MyErrors.UnauthenticatedUserException(e.details())
        if e.code() == StatusCode.UNAVAILABLE:
            raise MyErrors.ConnectionFailedException(e.details())
        raise e
    

def execAction(requestId : str, requestType : Method, fileName : str, stub : FileServiceStub) -> bool:
    switch : dict = {
        Method.GET : downloadFile,
        Method.PUT : uploadFile,
        Method.DEL : deleteFile
    }
    action = switch[requestType]
    Debug.debug(f"Invio della richiesta di '{requestType.name} {fileName}' all'edge peer")
    # Esecuzione dell'azione
    result = action(fileName, requestId, stub)
    return result


def downloadFile(filename : str, requestId : str, stub : FileServiceStub) -> bool:
    try:
        # Otteniamo i chunks dalla chiamata gRPC
        chunks = stub.Download(FileDownloadRequest(requestId = requestId, fileName = filename))
        # Scriviamo i chunks in un file e ritorniamo l'esito della scrittura
        return FileService().writeChunks(chunks = chunks, filename = filename)
    except StopIteration as e:
        raise MyErrors.RequestFailedException("Errore durante la lettura dei chunks ricevuti.")
    except grpc.RpcError as e:
        manageGRPCError(e)


def deleteFile(fileName : str, requestId : str, stub : FileServiceStub) -> bool:
    # Richiesta delete del file
    try:
        response : FileResponse = stub.Delete(FileDeleteRequest(fileName = fileName, requestId = requestId))
    except grpc.RpcError as e:
        if e.code() == StatusCode.UNKNOWN:
            grpcCustomError = json.loads(e.details())
            match grpcCustomError["ErrorCode"]:
                case ErrorCodes.S3_ERROR:
                    raise MyErrors.RequestFailedException("Fallimento su S3.\r\n" + grpcCustomError["ErrorMessage"])    
            raise MyErrors.UnknownException("Errore nell'operazione. Non si hanno maggiori dettagli su cosa è andato storto.\r\n" + e.details()) 
        if e.code() == StatusCode.UNAVAILABLE:
            raise MyErrors.ConnectionFailedException("Connessione con il server fallita.\r\n" + e.details())
    return response.success


def uploadFile(filename : str, requestId : str, stub : FileServiceStub) -> bool :
    response : BalancerResponse
    try:
        # Otteniamo la size del file da inviare
        fileSize = os.path.getsize(filename = os.environ.get("FILES_PATH") + filename)
        # Dividiamo il file in chunks
        chunks = FileService().getChunks(filename = filename)
        # Effettuiamo la chiamata gRPC
        response = stub.Upload(
            chunks,
            metadata = (('file_name', buildUploadFileName(filename)), ('request_id', requestId), ('file_size', str(fileSize)), )
        )
    except IOError as e:
            raise MyErrors.FailedToOpenException(f"Impossibile aprire file: {str(e)}")
    except grpc.RpcError as e:
        manageGRPCError(e)
    return response.success


def buildUploadFileName(fileName:str) -> str:
    global userInfo
    # Split del nome del file in base ed estensione (se presente)
    parts = fileName.split('.')
    # Se l'estensione del file non è presente, prendiamo l'intero nome del file come base
    if len(parts) == 1:
        base_name = parts[0]
        extension = ''
    else:
        base_name = '.'.join(parts[:-1])
        extension = parts[-1]
    # Concatena il filename iniziale, lo username e l'estensione (se presente)
    uploadFileName = f"{base_name}_{userInfo.username}.{extension}" if extension else f"{base_name}_{userInfo.username}"
    return uploadFileName


def manageGRPCError(e):
    if e.code() == StatusCode.UNKNOWN:
        grpcCustomError = json.loads(e.details())
        match grpcCustomError["ErrorCode"] :
            case ErrorCodes.FILE_NOT_FOUND_ERROR:
                raise MyErrors.FileNotFoundException("File richiesto non trovato.\r\n" + grpcCustomError["ErrorMessage"])
            case ErrorCodes.INVALID_METADATA:
                raise MyErrors.InvalidMetadataException("Metadata inviati non validi.\r\n" + grpcCustomError["ErrorMessage"])
            case ErrorCodes.FILE_READ_ERROR:
                raise MyErrors.RequestFailedException("Fallimento del server sulla lettura del file.\r\n" + grpcCustomError["ErrorMessage"])
            case ErrorCodes.S3_ERROR:
                raise MyErrors.RequestFailedException("Fallimento su S3.\r\n" + grpcCustomError["ErrorMessage"])    
        raise MyErrors.UnknownException("Errore sconosciuto nell'operazione.\r\n" + e.details()) 
    if e.code() == StatusCode.UNAVAILABLE:
        raise MyErrors.ConnectionFailedException("Connessione con il server fallita.\r\n" + e.details())
    raise e

def login(username : str, passwd : str) -> bool :
    global userInfo
    try:
        channel = grpc.insecure_channel(os.environ.get("LOAD_BALANCER_ADDR"))
        stub = BalancingServiceStub(channel)
        userInfo = User(username=username, passwd=passwd)
        loginResponse : LoginResponse = stub.LoginClient(userInfo)
        return loginResponse.logged
    except grpc.RpcError as e:
        if e.code() == StatusCode.UNAVAILABLE:
            raise MyErrors.ConnectionFailedException("Connessione con il balancer fallita.\r\n" + e.details())
        raise e

class FileService:

    def getChunks(self, filename):
            try:
                with open(os.environ.get("FILES_PATH") + filename, "rb") as file :
                    chunkSize = int(os.environ.get("CHUNK_SIZE"))
                    chunk = file.read(chunkSize)            
                    while chunk:
                        Debug.debug("Inviato un chunk di " + str(len(chunk)) + " bytes")
                        fileChunk : FileChunk = FileChunk(chunk = chunk)
                        yield fileChunk
                        chunk = file.read(chunkSize)

            except Exception as e:
                print(e)

    def readFile(self, file : io.BufferedReader):
        chunkSize = int(os.environ.get("CHUNK_SIZE"))
        chunk = file.read(chunkSize)
        count = 0
        print(count)
        while chunk:
            fileChunk : FileChunk = FileChunk(chunk = chunk)
            count+=1
            print(count)
            yield fileChunk
            chunk = file.read(chunkSize)

    def writeChunks(self, filename : str, chunks):
        return self.downloadFile(chunk_list = chunks, filename = filename)
        
    def downloadFile(self, filename : str, chunk_list) -> bool:
        chunk : FileChunk = next(chunk_list)
        try:
            with open(os.environ.get("FILES_PATH") + filename, "wb") as file:
                file.write(chunk.chunk)
                for chunk in chunk_list:
                    file.write(chunk.chunk)
        except IOError as e:
            raise MyErrors.FailedToOpenException(f"Impossibile aprire file: {str(e)}")

        return True