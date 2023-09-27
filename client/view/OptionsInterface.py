from controller import Controller
from engineering.Method import Method
from engineering import MyErrors, Debug
import getpass
from utils.Utils import *

def perform_action(request_type : Method, file_name : str):
    colored_print(f"La richiesta '{request_type.name} {file_name}' è in elaborazione.", Color.YELLOW)
    
    success : bool = Controller.sendRequestForFile(requestType = request_type, fileName = file_name)

    if not success:
        colored_print(f"La richiesta '{request_type.name} {file_name}' è fallita.", Color.RED)
        return

    colored_print(f"La richiesta '{request_type.name} {file_name}' è stata soddisfatta.", Color.GREEN)

def main():
    displayLoginBanner()
    loginSuccessful : bool = False
    username : str
    reset : int = 0
    while not loginSuccessful:
        username = input("\033[33mUSERNAME: \033[0m").strip()
        password = getpass.getpass("\033[33mPASSWORD: \033[0m").strip()
        try:
            loginSuccessful = Controller.login(username=username, passwd=password)
            if not loginSuccessful:
                reset += 1
                colored_print(f"Wrong username or password. Try again!", Color.RED)
            if reset == 3:
                reset = 0
                displayLoginBanner()
        except MyErrors.ConnectionFailedException  as e:
            reset += 1
            colored_print(f"Ci sono stati errori durante la connessione con il server. Ritenta.", Color.RED)
            Debug.errorDebug(e.message)
    option_interface(username)

def printError(message : str, e : Exception = None):
    colored_print(message, Color.RED)
    if e != None:
        Debug.errorDebug(e.message)

def option_interface(username:str):
    displayMenuBanner(username)

    while True:
        colored_print("Inserisci l'azione (get/put/delete/clear/exit) >>> ", Color.YELLOW, end = "")
        action = input("").strip().lower()
        if action == "exit":
            clearScreen()
            return
        elif action == "clear":
            displayMenuBanner(username)
        elif action in ["get", "put", "delete"]:
            while True:
                colored_print("Inserisci il nome del file >>> ", Color.YELLOW, end = "")
                file_name = input("").strip()
                if not file_name:
                    printError(None, "Non puoi richiedere un file con nome vuoto. Riprova inserendo il nome.", Color.RED)
                else:
                    break
            try:
                if action == "get":
                    perform_action(Method.DOWNLOAD, file_name)
                elif action == "put":
                    perform_action(Method.UPLOAD, file_name)
                elif action == "delete":
                    perform_action(Method.DELETE, file_name)
            except (MyErrors.InvalidMetadataException, MyErrors.ConnectionFailedException)  as e:
                printError(f"Ci sono stati errori durante la connessione con il server. Ritenta.", e)
            except MyErrors.RequestFailedException as e:
                printError("Il server non è riuscito a soddisfare la richiesta a causa di qualche errore. Ritenta.", e)
            except MyErrors.FileNotFoundException  as e:
                printError("File richiesto non trovato. Il nome del file è corretto?", e)
            except MyErrors.FailedToOpenException as e:
                printError("Impossibile aprire il file. Assicurati che il file esista.", e)
            except MyErrors.NoServerAvailableException as e:
                printError("Nessun server è disponibile. Ritenta più tardi.", e)
            except MyErrors.LocalFileNotFoundException as e:
                printError("Il file non esiste in locale.", e)
            except MyErrors.UnauthenticatedUserException as e:
                printError("Utente non autorizzato. Impossibile procedere con la richiesta.", e)
            except MyErrors.UnknownException as e:
                printError("Errore non identificato. La richiesta non è stata soddisfatta.", e)
            except TimeoutError as e:
                printError("Per troppo tempo non è stata ricevuta risposta dal server. Richiesta fallita.", e)
            except MyErrors.DownloadFailedException as e:
                printError("Errore ricevuto durante la ricezione del file.", e)
            except MyErrors.S3Exception as e:
                printError("S3 non è riuscito a soddisfare la richiesta.", e)
        else:
            printError("Azione non valida. Riprova.")

        print("")

