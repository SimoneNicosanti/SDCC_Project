from engineering import RabbitSingleton, Debug
from engineering.Ticket import Ticket, Method
from proto.file_transfer.File_pb2 import *
from proto.file_transfer.File_pb2_grpc import *
from utils import Utils
import json, grpc, os, pika.channel, pika

ticket_id_list : list = []

def sendRequestForFile(requestType : Method, fileName : str) -> bool:

    # Otteniamo l'indirizzo del peer da contattare
    ticket : Ticket = getTicket()
    if ticket == None:
        return None
    ticket_id_list.append(ticket.ticket_id)

    channel = grpc.insecure_channel(ticket.peer_addr)
    stub = FileServiceStub(channel)

    return execAction(ticket_id = ticket.ticket_id, requestType = requestType, filename = fileName, stub = stub)

def callback(ch, method, properties, body):
    data = json.load(body)
    serverEndpoint = data['ServerEndpoint']
    ticket_id = data['Id']
    ticket = Ticket(serverEndpoint, ticket_id)
    print(ticket)


def getTicket() -> Ticket:

    queue_name = os.environ.get("QUEUE_NAME")
    channel : pika.channel.Channel = RabbitSingleton.getRabbitChannel()
    channel.queue_declare(queue = queue_name, durable=True)
    
    while True :
        value = channel.basic_consume(queue_name, on_message_callback = callback, auto_ack=True)
        print(value)
    
#import pika

# Set up connection parameters
#connection_params = pika.ConnectionParameters('localhost')
#connection = pika.BlockingConnection(connection_params)
#channel = connection.channel()

# Declare a queue
#queue_name = 'my_queue'
#channel.queue_declare(queue=queue_name)

#while True:
#    method_frame, header_frame, body = channel.basic_get(queue=queue_name)
#    
#    if method_frame:
#        print(f"Received message: {body.decode()}")
#        channel.basic_ack(delivery_tag=method_frame.delivery_tag)
#    else:
#        print("No messages in the queue")
#        break
#
#connection.close()



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
    
    # Rimuoviamo il ticket dalla lista di ticket_id che stiamo usando
    ticket_id_list.remove(ticket_id)
    
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