import csv, os

def writeTimeOnFile(outputFileName: str, timeValue : float, requestedFileName : str, requestType : str):
    filePath = os.path.join("/Results", outputFileName)
    
    if not os.path.exists(filePath):
        with open(filePath, "x+") as file:
            writer = csv.writer(file)
            writer.writerow(["RequestType", "RequestedFileName", "FileSize [MB]", "Time [s]"])

    with open(os.path.join("/Results", outputFileName), "a") as file:
        writer = csv.writer(file)
        if requestType == "DELETE":
            fileSize = "NaN"
        else :
            byteFileSize = os.path.getsize(filename = os.environ.get("FILES_PATH") + requestedFileName)
            fileSize = float(byteFileSize) / 1048576.0
        writer.writerow([requestType, requestedFileName, fileSize, timeValue])