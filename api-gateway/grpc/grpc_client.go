package grpc

import (
	"context"
	"fmt"
	"log"

	authpb "github.com/Dattt2k2/golang-project/auth-service/gRPC/service"
	"google.golang.org/grpc"
)

var authClient authpb.AuthServiceClient

func InitGrpcClient(address string){
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil{
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	authClient = authpb.NewAuthServiceClient(conn)
	log.Println("Connected to gRPC server")
}

func VerifyToken(token string)(*authpb.VerifyTokenResponse, error){
	req  := &authpb.VerifyTokenRequest{Token: token}
	res, err := authClient.VerifyToken(context.Background(), req)
	if err != nil{
		return nil, fmt.Errorf("Failed to verify token: %v", err)
	}
	return res, nil
}