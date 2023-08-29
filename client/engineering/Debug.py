from utils.Utils import *

def debug(debugString : str) -> None:
    colored_print("[*DEBUG*] -> " + debugString, Color.CYAN)

def errorDebug(debugString : str) -> None:
    colored_print("[*DEBUG*] -> " + debugString, Color.MAGENTA)