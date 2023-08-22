# from view import LoginInterface
from controller import Controller


# def proto_attempt() :
#     channel = grpc.insecure_channel("login:50051")
#     client_stub = LoginServiceStub(channel)
    
#     userInfo : UserInfo = UserInfo(username = "bella", passwd = "a", email = "tutti")
#     response : Response = client_stub.login(userInfo)
#     print(response.response)

if __name__ == "__main__":
    # LoginInterface.startup_login_interface()
    Controller.getFile("file_prova.txt")