from controller import Controller

def perform_get(file_name):
    print(f"DOWNLOADING: {file_name}")
    
    # Implementazione per l'azione GET
    Controller.getFile(fileName = file_name)

    print(f"DOWNLOADED: {file_name}")

def perform_put(file_name):
    print(f"UPLOADING: {file_name}")

    # Implementazione per l'azione PUT
    Controller.putFile(fileName = file_name)
    
    print(f"UPLOADED: {file_name}")

def perform_delete(file_name):
    print(f"DELETING: {file_name}")

    # Implementazione per l'azione DELETE
    Controller.deleteFile(fileName = file_name)
    
    print(f"DELETED: {file_name}")

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

