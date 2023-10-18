from engineering import Debug, MyErrors
from proto.file_transfer.FileTransfer_pb2 import *
from proto.file_transfer.FileTransfer_pb2_grpc import *
import os, io

def getChunks(filename):
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

def readFile(file : io.BufferedReader):
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

def writeFile(filename : str, chunk_list) -> bool:
    try:
        with open(os.environ.get("FILES_PATH") + filename, "wb") as file:
            writeChunks(file, chunk_list)
    except IOError as e:
        raise MyErrors.FailedToOpenException(f"Impossibile aprire file: {str(e)}")

    return True

def writeChunks(file, chunk_list):
    seqNum = 0
    for chunk in chunk_list:
        Debug.debug("Ricevuto un chunk di " + str(len(chunk.chunk)) + " bytes")
        if chunk.seqNum != seqNum:
            if chunk.seqNum == 0:
                file.seek(0, 0)
                seqNum = 0
            else:
                raise MyErrors.DownloadFailedException("Errore durante il download del file.")
        file.write(chunk.chunk)
        seqNum+=1