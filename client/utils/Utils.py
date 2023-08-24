import jproperties

def readProperties(fileName : str, propertyName : str) -> str :
    configs = jproperties.Properties()
    with open(fileName, "rb") as propertiesFile :
        configs.load(propertiesFile)
        return configs.get(propertyName).data