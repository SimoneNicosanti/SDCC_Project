import pika
import grpc
# from proto.Login_pb2_grpc import LoginServiceStub
# from proto.Login_pb2 import UserInfo, Response
from controller import Controller


# def proto_attempt() :
#     channel = grpc.insecure_channel("login:50051")
#     client_stub = LoginServiceStub(channel)
    
#     userInfo : UserInfo = UserInfo(username = "bella", passwd = "a", email = "tutti")
#     response : Response = client_stub.login(userInfo)
#     print(response.response)

def main() :
    Controller.getFile("La Ginestra")




if __name__ == "__main__" :
    main()