from controller import Controller
from engineering.Method import Method
import os

def RunTest(argv : list[str]) :
    username : str = argv[argv.index('-u') + 1]
    password : str = argv[argv.index("-p") + 1]

    operation : str = argv[argv.index("-o") + 1]
    fileName : str = argv[argv.index("-f") + 1]

    edgeNum : str = argv[argv.index("-e") + 1]
    os.environ.__setitem__("EDGE_NUM", edgeNum)

    outputFile : str = argv[argv.index("-w") + 1]
    writeTime : bool = (outputFile != "NO")
    os.environ.__setitem__("OUTPUT_FILE", outputFile)

    Controller.userInfo.username = username
    Controller.userInfo.passwd = password

    if operation == "download" :
        method = Method.DOWNLOAD
    elif operation == "upload" :
        method = Method.UPLOAD
    elif operation == "delete" :
        method = Method.DELETE
    else :
        return
    
    Controller.sendRequestForFile(method, fileName, writeTime = writeTime)