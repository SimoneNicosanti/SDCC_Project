from controller import Controller
from engineering.Ticket import Method
from engineering import MyErrors, Debug
from utils.Utils import *

def perform_action(request_type : Method, file_name : str):
    colored_print(f"Richiesta '{request_type.name} {file_name}' in elaborazione.", Color.YELLOW)
        
    success : bool = Controller.sendRequestForFile(requestType = request_type, fileName = file_name)
    

    if not success:
        colored_print(f"La richiesta '{request_type.name} {file_name}' e' fallita.", Color.RED)
        return

    colored_print(f"Richiesta '{request_type.name} {file_name}' soddisfatta.", Color.GREEN)

def main():
    colored_print("Benvenuto nella CLI di gestione file!", Color.YELLOW)

    while True:
        colored_print("Inserisci l'azione (get/put/delete) >>> ", Color.YELLOW, end = "")
        action = input("").strip().lower()

        if action in ["get", "put", "delete"]:
            colored_print("Inserisci il nome del file >>> ", Color.YELLOW, end = "")
            file_name = input("").strip()

            try:
                if action == "get":
                    perform_action(Method.GET, file_name)
                elif action == "put":
                    perform_action(Method.PUT, file_name)
                elif action == "delete":
                    perform_action(Method.DEL, file_name)
            except (MyErrors.InvalidTicketException, MyErrors.ConnectionFailedException)  as e:
                colored_print(f"Ci sono stati errori durante la connessione con il server. Ritenta.", Color.RED)
                Debug.errorDebug(e.message)
            except MyErrors.RequestFailedException as e:
                colored_print("Il server non è riuscito a soddisfare la risorsa a causa di qualche errore. Ritenta.", Color.RED)
                Debug.errorDebug(e.message)
            except MyErrors.FileNotFoundException  as e:
                colored_print("File richiesto non trovato. Il nome del file è corretto?", Color.RED)
                Debug.errorDebug(e.message)
            except MyErrors.FailedToOpenException as e:
                colored_print("Impossibile aprire il file. Assicurati che il file esista.", Color.RED)
                Debug.errorDebug(e.message)
            except MyErrors.NoServerAvailable as e:
                colored_print("Nessun server è disponibile. Ritenta più tardi.", Color.RED)
                Debug.errorDebug(e.message)
            except MyErrors.FileNotFound as e:
                colored_print("Il file non esiste in locale.", Color.RED)
                Debug.errorDebug(e.message)
        else:
            colored_print("Azione non valida. Riprova.", Color.RED)

        print("")

