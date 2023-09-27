import os, jproperties

class Color:
    RESET = '\033[0m'
    BLACK = '\033[30m'
    RED = '\033[31m'
    GREEN = '\033[32m'
    YELLOW = '\033[33m'
    BLUE = '\033[34m'
    MAGENTA = '\033[35m'
    CYAN = '\033[36m'
    WHITE = '\033[37m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

def colored_print(text, color, end = "\n"):
    print(f"{color}{text}{Color.RESET}", end = end)

def readProperties(fileName : str, propertyName : str) -> str :
    configs = jproperties.Properties()
    with open(fileName, "rb") as propertiesFile :
        configs.load(propertiesFile)
        return configs.get(propertyName).data

def clearScreen():
    if os.name == 'posix':
        os.system('clear')
    elif os.name == 'nt':
        os.system('cls')

def displayLoginBanner():
    clearScreen()
    colored_print("""

██╗       ██████╗   ██████╗  ██╗ ███╗   ██╗
██║      ██╔═══██╗ ██╔════╝  ██║ ████╗  ██║
██║      ██║   ██║ ██║  ███╗ ██║ ██╔██╗ ██║
██║      ██║   ██║ ██║   ██║ ██║ ██║╚██╗██║
███████╗ ╚██████╔╝ ╚██████╔╝ ██║ ██║ ╚████║
╚══════╝  ╚═════╝   ╚═════╝  ╚═╝ ╚═╝  ╚═══╝
                                    
                  
I N S E R I R E    L E    C R E D E N Z I A L I
                                    
    """,Color.YELLOW)

def displayMenuBanner(username:str):
    clearScreen()
    colored_print(f"""
███████╗  █████╗  ███████╗
██╔════╝ ██╔══██╗ ██╔════╝
███████╗ ███████║ █████╗  
╚════██║ ██╔══██║ ██╔══╝  
███████║ ██║  ██║ ███████╗
╚══════╝ ╚═╝  ╚═╝ ╚══════╝    Storage nel Cloud Continuum

                        
B E N T O R N A T O   {addSpacesBetweenChars(username.upper())} !
                  
    """,Color.YELLOW)

def addSpacesBetweenChars(input_string:str) -> str:
    spaced_string = ' '.join(input_string)
    return spaced_string

def buildUploadFileName(fileName:str, username:str) -> str:
    # Split del nome del file in base ed estensione (se presente)
    parts = fileName.split('.')
    # Se l'estensione del file non è presente, prendiamo l'intero nome del file come base
    if len(parts) == 1:
        base_name = parts[0]
        extension = ''
    else:
        base_name = '.'.join(parts[:-1])
        extension = parts[-1]
    # Concatena il filename iniziale, lo username e l'estensione (se presente)
    uploadFileName = f"{base_name}_{username}.{extension}" if extension else f"{base_name}_{username}"
    return uploadFileName