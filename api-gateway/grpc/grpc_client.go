package grpcClient

import (
	"context"
	"sync"

	"github.com/Dattt2k2/golang-project/api-gateway/logger"
	authpb "github.com/Dattt2k2/golang-project/auth-service/gRPC/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// var authClient authpb.AuthServiceClient

// func InitGrpcClient(address string){
// 	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil{
// 		log.Fatalf("Failed to connect to gRPC server: %v", err)
// 	}
// 	authClient = authpb.NewAuthServiceClient(conn)
// 	log.Println("Connected to gRPC server")
// }

// func VerifyToken(token string)(*authpb.VerifyTokenResponse, error){
// 	req  := &authpb.VerifyTokenRequest{Token: token}
// 	res, err := authClient.VerifyToken(context.Background(), req)
// 	if err != nil{
// 		return nil, fmt.Errorf("Failed to verify token: %v", err)
// 	}
// 	return res, nil
// }


var (
	authClient authpb.AuthServiceClient
	once 	 sync.Once
	clientErr error
)


func InitGrpcClient(address string) error{
	once.Do(func(){
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil{
			logger.Err("Failed to connect to gRPC server", err)
			clientErr = err
			return
		}

		authClient = authpb.NewAuthServiceClient(conn)
	})

	return clientErr
}


func VerifyToken(token string) (*authpb.VerifyTokenResponse, error){
	req := &authpb.VerifyTokenRequest{
		Token: token,
	}

	return authClient.VerifyToken(context.Background(), req)
}