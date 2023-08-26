from engineering import RabbitSingleton, Debug
from engineering.Ticket import Ticket, Method
from proto.file_transfer.File_pb2 import *
from proto.file_transfer.File_pb2_grpc import *
from asyncio import Semaphore
from pika.adapters.blocking_connection  import BlockingChannel
from utils import Utils
import json, grpc, os

ticket_id_list : list = []
data = None
sem : Semaphore = Semaphore()

def sendRequestForFile(requestType : Method, fileName : str) -> bool:

    # Otteniamo l'indirizzo del peer da contattare
    ticket : Ticket = getTicket()
    if ticket == None:
        return False
    ticket_id_list.append(ticket.ticket_id)

    # Preparazione chiamata gRPC
    channel = grpc.insecure_channel(ticket.peer_addr)
    stub = FileServiceStub(channel)

    # Esecuzione dell'azione
    result = execAction(ticket_id = ticket.ticket_id, requestType = requestType, filename = fileName, stub = stub)


    # Rimuoviamo il ticket dalla lista di ticket_id che stiamo usando
    ticket_id_list.remove(ticket.ticket_id)

    return result

def callback(ch : BlockingChannel, method, properties, body):
    global data 
    if method == None:
        data = None
    else:
        data = json.load(body)
    
    ch.stop_consuming()


def getTicket() -> Ticket:

    queue_name = os.environ.get("QUEUE_NAME")
    channel : BlockingChannel = RabbitSingleton.getRabbitChannel()
    channel.queue_declare(queue = queue_name, durable = True)
    
    channel.basic_consume(queue = queue_name, on_message_callback = callback, auto_ack = True)
    # All'interno della callback viene scritto il contenuto del messaggio in 'data'
    sem.acquire()
    channel.start_consuming()
    global data
    ticket = None
    if data != None:
        serverEndpoint = data['ServerEndpoint']
        ticket_id = data['Id']
        ticket = Ticket(serverEndpoint, ticket_id)
        Debug.debug(ticket)
    # Dopo aver creato il Ticket possiamo rilasciare il semaforo
    sem.release()

    return ticket



def execAction(ticket_id : str, requestType : Method, filename : str, stub : FileServiceStub) -> bool:

    switch : dict = {
        Method.GET : getFile,
        Method.PUT : putFile,
        Method.DEL : deleteFile
    }

    action = switch(requestType)

    Debug.debug(f"Invio della richiesta di '{requestType.name} {filename}' all'edge peer")

    # Esecuzione dell'azione
    result = action(filename, ticket_id, stub)
    
    return result

def getFile(filename : str, ticket_id : str, stub : FileServiceStub) -> bool :
    # Otteniamo i chunks dalla chiamata gRPC
    chunks = stub.Download(FileDownloadRequest(ticket_id = ticket_id, file_name = filename))

    # Scriviamo i chunks in un file e ritorniamo l'esito della scrittura
    return FileService().writeChunks(ticket_id = ticket_id, chunks = chunks)


def putFile(filename : str, ticket_id : str, stub : FileServiceStub) -> bool :
    # Dividiamo il file in chunks
    chunks = FileService().getChunks(ticket_id = ticket_id, filename = filename)

    # Effettuiamo la chiamata gRPC 
    response : Response = stub.Upload(chunks)

    # Se la risposta non riguarda la nostra richiesta, lanciamo una eccezione
    if response.ticket_id != ticket_id:
        raise Exception()
    
    return response.success


def deleteFile(fileName : str) -> bool:
    #TODO
    pass

def login(username : str, passwd : str, email : str) -> bool :
    if username == "sae" and passwd == "admin" and email == "admin@sae.com":
        return True


class FileService:
    global ticket_id_list

    def getChunks(self, ticket_id : str, filename : str):
        #Security check
        if ticket_id not in ticket_id_list:
            Debug.debug(f"[*ERROR*] - Security check failed --> {ticket_id} is not in the opened ticket list")
            return

        try:
            with open("/files/" + filename, "rb") as file :
                chunkSize = int(Utils.readProperties("conf.properties", "CHUNK_SIZE"))
                chunk = file.read(chunkSize)
                while chunk:
                    fileChunk : FileChunk = FileChunk(ticket_id = ticket_id, file_name = filename, chunk = chunk)
                    yield fileChunk
                    chunk = file.read(chunkSize)
        except IOError as e:
            Debug.debug(f"[*ERROR*] - Couldn't open the file: {str(e)}")

    def writeChunks(self, ticket_id : str, chunks):
        try:
            return self.downloadFile(ticket_id, chunks)
        except IOError as e:
            Debug.debug(f"[*ERROR*] - {str(e)}")
            return False

    def downloadFile(ticket_id : str, chunk_list) -> bool:

        chunk : FileChunk = next(chunk_list)
        filename = chunk.file_name
        try:
            with open("/files/" + filename, "x") as file:
                file.write(chunk.chunk)
                for chunk in chunk_list:
                    file.write(chunk.chunk)

        except IOError as e:
            Debug.debug(f"[*ERROR*] - Couldn't open the file: {str(e)}")
            return False

        return True