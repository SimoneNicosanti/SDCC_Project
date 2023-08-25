from controller import Controller
from engineering.Message import Method

def perform_get(file_name):
    print(f"DOWNLOADING: {file_name}")
    
    # Implementazione per l'azione GET
    Controller.sendRequestForFile(Method.GET, fileName = file_name)

    print(f"DOWNLOADED: {file_name}")

def perform_put(file_name):
    print(f"UPLOADING: {file_name}")

    # Implementazione per l'azione PUT
    Controller.sendRequestForFile(Method.PUT, fileName = file_name)
    
    print(f"UPLOADED: {file_name}")

def perform_delete(file_name):
    print(f"DELETING: {file_name}")

    # Implementazione per l'azione DELETE
    Controller.sendRequestForFile(Method.DEL, fileName = file_name)
    
    print(f"DELETED: {file_name}")

def 

def main():
    print("Benvenuto nella CLI di gestione file!")

    while True:
        action = input("Inserisci l'azione (get/put/delete): ").strip().lower()

        if action in ["get", "put", "delete"]:
            file_name = input("Inserisci il nome del file: ").strip()

            if action == "get":
                perform_get(file_name)
            elif action == "put":
                perform_put(file_name)
            elif action == "delete":
                perform_delete(file_name)

            break
        else:
            print("Azione non valida. Riprova.")

