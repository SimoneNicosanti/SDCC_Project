import csv, os, fcntl

MB = 1024 * 1024

def writeTimeOnFile(timeValue : float, requestedFileName : str, requestType : str):
    outputFileName : str = os.environ.get("OUTPUT_FILE")
    
    filePath = os.path.join("/Results", outputFileName)
    if not os.path.exists(filePath):
        with open(filePath, "x+") as file:
            writer = csv.writer(file)
            writer.writerow(["RequestType", "FileSize [MB]", "Time [s]", "EdgeNum"])

    with open(os.path.join("/Results", outputFileName), "a") as file:
        writer = csv.writer(file)
        if requestType == "DELETE":
            fileSize = "NaN"
        else :
            byteFileSize = os.path.getsize(filename = os.environ.get("FILES_PATH") + requestedFileName)
            fileSize = float(byteFileSize) / MB
        
        fcntl.flock(file, fcntl.LOCK_EX)
        writer.writerow([requestType, fileSize, timeValue, os.environ.get("EDGE_NUM")])