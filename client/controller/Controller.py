from engineering import Debug, MyErrors
from engineering.Method import Method
from proto.file_transfer.FileTransfer_pb2 import *
from proto.file_transfer.FileTransfer_pb2_grpc import *
from proto.load_balancer.LoadBalancer_pb2 import *
from proto.load_balancer.LoadBalancer_pb2_grpc import *
from utils import Utils
from grpc import StatusCode
from asyncio import Semaphore
from controller import FileServiceController, S3Controller
from dao import CSVWriter
import json, grpc, os, time

data = None
userInfo : User = User(username="a",passwd="a")
sem : Semaphore = Semaphore()

def sendRequestForFile(requestType : Method, fileName : str, writeTime : bool = False) -> bool:
    global userInfo
    if requestType == Method.UPLOAD and not os.path.exists(os.environ.get("FILES_PATH") + fileName):
        raise MyErrors.LocalFileNotFoundException("Il file da caricare non esiste in locale.")
    # Otteniamo l'indirizzo del peer da contattare
    try:
        response: BalancerResponse = getEdgeFromBalancer()
    except MyErrors.NoServerAvailableException as e:
        startTime = time.time()
        S3Controller.serveRequestDirectlyFromS3(requestType=requestType, fileName=fileName, username=userInfo.username)
        endTime = time.time()
        if (writeTime) :
            CSVWriter.writeTimeOnFile("S3_Direct.csv", endTime - startTime, fileName, requestType.name)
        return True

    # Preparazione chiamata gRPC
    try:
        channel : grpc.Channel = grpc.insecure_channel(response.edgeIpAddr, options=[('grpc.max_receive_message_length', int(os.environ.get("MAX_GRPC_MESSAGE_SIZE")))])
        stub = FileServiceStub(channel)
    except grpc.RpcError as e:
        if e.code() == StatusCode.UNAVAILABLE:
            raise MyErrors.ConnectionFailedException(e.details())
        raise e
    # Esecuzione dell'azione
    startTime = time.time()
    response = execAction(requestId = response.requestId, requestType = requestType, fileName = fileName, stub = stub)
    endTime = time.time()
    if (writeTime) :
        CSVWriter.writeTimeOnFile("With_System.csv", endTime - startTime, fileName, requestType.name)
    return response

def getEdgeFromBalancer() -> BalancerResponse:
    try:
        channel = grpc.insecure_channel(os.environ.get("LOAD_BALANCER_ADDR"))
        stub = BalancingServiceStub(channel)
        return stub.GetEdge(userInfo)
    except grpc.RpcError as e:
        if e.code() == StatusCode.UNKNOWN:
            grpcCustomError = json.loads(e.details())
            match grpcCustomError["ErrorCode"]:
                case ErrorCodesLoadBalancer.NO_SERVER_AVAILABLE:
                    Debug.debug("Nessun server disponibile. Ripiego su S3.")
                    raise MyErrors.NoServerAvailableException("")
            raise MyErrors.UnknownException("Errore nell'operazione. Non si hanno maggiori dettagli su cosa è andato storto.\r\n" + e.details()) 
        if e.code() == StatusCode.UNAUTHENTICATED:
            raise MyErrors.UnauthenticatedUserException(e.details())
        if e.code() == StatusCode.UNAVAILABLE:
            raise MyErrors.ConnectionFailedException(e.details())
        raise e
    

def execAction(requestId : str, requestType : Method, fileName : str, stub : FileServiceStub) -> bool:
    switch : dict = {
        Method.DOWNLOAD : downloadFile,
        Method.UPLOAD : uploadFile,
        Method.DELETE : deleteFile
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
        return FileServiceController.writeFile(chunk_list = chunks, filename = filename)
    except StopIteration as e:
        raise MyErrors.RequestFailedException("Errore durante la lettura dei chunks ricevuti.")
    except grpc.RpcError as e:
        if os.path.exists(os.environ.get("FILES_PATH") + filename):
            os.remove(os.environ.get("FILES_PATH") + filename)
            Debug.debug(f"Il file {filename} è stato eliminato.")
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
                case ErrorCodes.FILE_NOT_FOUND_ERROR:
                    raise MyErrors.FileNotFoundException("Il file non esiste o è già stato eliminato da S3.\r\n" + grpcCustomError["ErrorMessage"])
            raise MyErrors.UnknownException("Errore nell'operazione. Non si hanno maggiori dettagli su cosa è andato storto.\r\n" + e.details()) 
        if e.code() == StatusCode.UNAVAILABLE:
            raise MyErrors.ConnectionFailedException("Connessione con il server fallita.\r\n" + e.details())
    return response.success


def uploadFile(filename : str, requestId : str, stub : FileServiceStub) -> bool :
    global userInfo
    response : BalancerResponse
    try:
        # Otteniamo la size del file da inviare
        fileSize = os.path.getsize(filename = os.environ.get("FILES_PATH") + filename)
        # Dividiamo il file in chunks
        chunks = FileServiceController.getChunks(filename = filename)
        # Effettuiamo la chiamata gRPC
        response = stub.Upload(
            chunks,
            metadata = (('file_name', Utils.buildUploadFileName(filename, userInfo.username)), ('request_id', requestId), ('file_size', str(fileSize)), )
        )
    except IOError as e:
        raise MyErrors.FailedToOpenException(f"Impossibile aprire file: {str(e)}")
    except grpc.RpcError as e:
        manageGRPCError(e)
    return response.success


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