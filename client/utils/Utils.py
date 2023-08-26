import jproperties

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