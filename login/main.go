package main

import (
	"context"
	"fmt"
	"log"
	"login/proto"
	"net"

	// "google.golang.org/genproto/googleapis/cloud/orchestration/airflow/service/v1"
	"google.golang.org/grpc"
)

type LoginServer struct {
	proto.UnimplementedLoginServiceServer
}

func (s *LoginServer) Login(ctx context.Context, userInfo *proto.UserInfo) (*proto.Response, error) {
	fmt.Printf("Received message")
	// TODO implementare la logica di login
	return &proto.Response{Response: proto.Response_ERROR}, nil
}

func main() {
	lis, err := net.Listen("tcp", "login:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// var opts []grpc.ServerOption
	grpcServer := grpc.NewServer()
	loginServer := &LoginServer{}
	proto.RegisterLoginServiceServer(grpcServer, loginServer)
	fmt.Printf("Listening for connections...")
	grpcServer.Serve(lis)

}
