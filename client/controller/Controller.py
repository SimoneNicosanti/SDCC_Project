from engineering import RabbitSingleton, Debug, MyErrors
from engineering.Ticket import Ticket, Method
from proto.file_transfer.ClientFileTransfer_pb2 import *
from proto.file_transfer.ClientFileTransfer_pb2_grpc import *
from grpc import StatusCode
from asyncio import Semaphore
from pika.adapters.blocking_connection  import BlockingChannel
from pika.exceptions import StreamLostError, ChannelWrongStateError
import json, grpc, os, io

ticket_id_list : list = []
data = None
sem : Semaphore = Semaphore()

def sendRequestForFile(requestType : Method, fileName : str) -> bool:

    count = 0 
    if requestType == Method.PUT and not os.path.exists(os.environ.get("FILES_PATH") + fileName):
        raise MyErrors.FileNotFound("Il file da caricare non esiste in locale.")

    # Otteniamo l'indirizzo del peer da contattare
    while(True):
        try:
            ticket : Ticket = getTicket()
            if ticket == None:
                return False
            ticket_id_list.append(ticket.ticket_id)
            break
        except (StreamLostError, ChannelWrongStateError) as e:
            Debug.debug("Connessione con RabbitMQ caduta.")
            if count >= 3:
                Debug.debug("Impossibile stabilire la connesione con la coda.")
                raise MyErrors.UnableToConnectWithRabbit() 
            count+=1
            RabbitSingleton.startNewConnection()
            Debug.debug("Connessione con RabbitMQ ristabilità.")

    # Preparazione chiamata gRPC
    try:
        channel : grpc.Channel = grpc.insecure_channel(ticket.peer_addr, options=[('grpc.max_receive_message_length', int(os.environ.get("MAX_GRPC_MESSAGE_SIZE")))])

        stub = FileServiceStub(channel)
    except grpc.RpcError as e:
        print(e.code())
        print(e.code().value)
        print(e.details())
        if e.code().value[0] == StatusCode.UNAVAILABLE:
            raise MyErrors.ConnectionFailedException("Connessione con il server fallita. E' possibile che il server abbia avuto un guasto.")
        raise e
    # Esecuzione dell'azione
    result = execAction(ticket_id = ticket.ticket_id, requestType = requestType, filename = fileName, stub = stub)
    # Rimuoviamo il ticket dalla lista di ticket_id che stiamo usando
    ticket_id_list.remove(ticket.ticket_id)

    return result

def getTicket() -> Ticket:

    queue_name = os.environ.get("QUEUE_NAME")
    channel : BlockingChannel = RabbitSingleton.getRabbitChannel()
    
    _, _, body = channel.basic_get(queue=queue_name, auto_ack=True)
    
    if body == None:
        raise MyErrors.NoServerAvailable("Nessun server è pronto a soddisfare la richiesta.")

    data = json.loads(body)
    serverEndpoint = data['ServerEndpoint']
    ticket_id = data['Id']
    ticket = Ticket(serverEndpoint, ticket_id)
    Debug.debug(f"Ottenuto il ticket '{ticket_id}' sulla connessione {ticket.peer_addr}'")
    return ticket


def execAction(ticket_id : str, requestType : Method, filename : str, stub : FileServiceStub) -> bool:

    switch : dict = {
        Method.GET : getFile,
        Method.PUT : putFile,
        Method.DEL : deleteFile
    }

    action = switch[requestType]

    Debug.debug(f"Invio della richiesta di '{requestType.name} {filename}' all'edge peer")

    # Esecuzione dell'azione
    result = action(filename, ticket_id, stub)

    return result


def getFile(filename : str, ticket_id : str, stub : FileServiceStub) -> bool :
    try:
        # Otteniamo i chunks dalla chiamata gRPC
        chunks = stub.Download(FileDownloadRequest(ticket_id = ticket_id, file_name = filename))
        # Scriviamo i chunks in un file e ritorniamo l'esito della scrittura
        return FileService().writeChunks(chunks = chunks, filename = filename)
    except StopIteration as e:
        raise MyErrors.RequestFailedException("Errore durante la lettura dei chunks ricevuti.")
    except grpc.RpcError as e:
        print(e.code())
        print(e.code().value)
        print(e.details())
        if e.code().value[0] == ErrorCodes.FILE_NOT_FOUND_ERROR:
            raise MyErrors.FileNotFoundException("File richiesto non trovato.")
        if e.code().value[0] == ErrorCodes.INVALID_TICKET:
            raise MyErrors.InvalidTicketException("Ticket usato non valido.")
        if e.code().value[0] == ErrorCodes.FILE_READ_ERROR or e.code().value[0] == ErrorCodes.S3_ERROR:
            raise MyErrors.RequestFailedException("Fallimento durante il soddisfacimento della richiesta.")
        if e.code() == StatusCode.UNAVAILABLE:
            raise MyErrors.ConnectionFailedException("Connessione con il server fallita.")
        raise e


def putFile(filename : str, ticket_id : str, stub : FileServiceStub) -> bool :
    response : Response
       
    try:
        # Otteniamo la size del file da inviare
        file_size = os.path.getsize(filename = os.environ.get("FILES_PATH") + filename)
        # Dividiamo il file in chunks
        #with open(os.environ.get("FILES_PATH") + filename, "rb") as file :
        chunks = FileService().getChunks(filename = filename)
        # Effettuiamo la chiamata gRPC
        response = stub.Upload(
            chunks,
            metadata = (('file_name', filename), ('ticket_id', ticket_id), ('file_size', str(file_size)), )
        )
    except IOError as e:
            raise MyErrors.FailedToOpenException(f"Impossibile aprire file: {str(e)}")
    except grpc.RpcError as e:
        print(e.code())
        print(e.code().value)
        print(e.details())
        if e.code().value[0] == ErrorCodes.FILE_NOT_FOUND_ERROR:
            raise MyErrors.FileNotFoundException("File richiesto non trovato.")
        if e.code().value[0] == ErrorCodes.INVALID_TICKET:
            raise MyErrors.InvalidTicketException("Ticket usato non valido.")
        if e.code().value[0] == ErrorCodes.FILE_READ_ERROR:
            raise MyErrors.RequestFailedException("Fallimento durante il soddisfacimento della richiesta.")
        if e.code().value[0] == ErrorCodes.S3_ERROR:
            raise MyErrors.RequestFailedException("Fallimento durante il soddisfacimento della richiesta. Errore durante l'interazione con S3.")
        if e.code() == StatusCode.UNAVAILABLE:
            raise MyErrors.ConnectionFailedException("Connessione con il server fallita.")
        raise e
    

    # Se la risposta non riguarda la nostra richiesta, lanciamo una eccezione
    if response.ticket_id != ticket_id:
        raise MyErrors.RequestFailedException("Impossibile soddisfare la richiesta. E' stata ricevuta una risposta relativa ad un'altra richiesta.")
    
    return response.success

def deleteFile(fileName : str) -> bool:
    #TODO implementazione operazione di delete del file
    pass

def login(username : str, passwd : str, email : str) -> bool :
    if username == "sae" and passwd == "admin" and email == "admin@sae.com":
        return True

class FileService:
    global ticket_id_list

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