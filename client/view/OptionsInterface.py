from controller import Controller
from engineering.Ticket import Method

def perform_action(request_type : Method, file_name : str):
    print(f"Richiesta '{request_type.name} {file_name}' in elaborazione.")
    
    success : bool = Controller.sendRequestForFile(requestType = request_type, fileName = file_name)

    if not success:
        print(f"La richiesta '{request_type.name} {file_name}' e' fallita.") 
        return

    print(f"Richiesta '{request_type.name} {file_name}' soddisfatta.")

def main():
    print("Benvenuto nella CLI di gestione file!")

    while True:
        action = input("Inserisci l'azione (get/put/delete): ").strip().lower()

        if action in ["get", "put", "delete"]:
            file_name = input("Inserisci il nome del file: ").strip()

            if action == "get":
                perform_action(Method.GET, file_name)
            elif action == "put":
                perform_action(Method.PUT, file_name)
            elif action == "delete":
                perform_action(Method.DEL, file_name)

            break
        else:
            print("Azione non valida. Riprova.")

