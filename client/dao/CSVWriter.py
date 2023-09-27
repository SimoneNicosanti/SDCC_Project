import csv, os

def writeTimeOnFile(outputFileName: str, timeValue : float, requestedFileName : str) :
    filePath = os.path.join("/Results", outputFileName)
    
    if not os.path.exists(filePath) :
        with open(filePath, "wb") as file :
            writer = csv.writer(file)
            writer.writerow(["RequestedFileName", "FileSize", "Time"])

    with open(os.path.join("/Results", outputFileName), "wb") as file :
        writer = csv.writer(file)
        fileSize = os.path.getsize(filename = os.environ.get("FILES_PATH") + requestedFileName)
        writer.writerow(requestedFileName, fileSize, timeValue)