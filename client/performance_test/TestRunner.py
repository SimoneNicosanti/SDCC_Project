from controller import Controller
from engineering.Method import Method

def RunTest(argv : list[str]) :
    username : str = argv[argv.index('-u') + 1]
    password : str = argv[argv.index("-p") + 1]

    operation : str = argv[argv.index("-o") + 1]
    fileName : str = argv[argv.index("-f") + 1]

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
    
    Controller.sendRequestForFile(method, fileName, writeTime = True)