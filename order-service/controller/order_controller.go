package controller

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Dattt2k2/golang-project/order-service/kafka"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	cartpb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
)

func OrderFromCart() gin.HandlerFunc{
	return func(c *gin.Context){
		
	}
}

func OrderFromProduct() gin.HandlerFunc{
	return func(c *gin.Context){
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil{
			log.Printf("Failed to connect to gRPC server: %v", err)
			return
		}
		defer conn.Close()

		client:= cartpb.NewCartServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req:= &cartpb.CartRequest{
			UserId: c.Param("userId"),
		}

		resp, err := client.GetCartItems(ctx, req)
		if err != nil{
			log.Printf("Failed to get cart items: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		total := calculateAmount(resp)
		log.Printf("Total amount: %v", total)

		orderEvent:= kafka.PaymentOrder{
			UserId: req.UserId,
			Amount: total,
			Products: resp.Items,
		}

		if err := kafka.ProducePaymentOrder(orderEvent); err != nil{
			log.Printf("Failed to produce payment order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}


		c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully"})

	}

}

func OrderDirectly() gin.HandlerFunc{

	type OrderRequest struct{
		Userid string `json:"user_id"`
		Items []struct{
			ProductId string `json:"product_id"`
			Quantity int32 `json:"quantity"`
			Price float64 `json:"price"`
		} `json:"items"`
	}

	return func(c *gin.Context){
		var orderReq OrderRequest
		if err := c.ShouldBindJSON(&orderReq); err != nil{
			log.Printf("Failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var total float64
		for _, items := range orderReq.Items{
			total += float64(items.Quantity) * float64(items.Price)
		}

		orderEvent:= kafka.PaymentOrder{
			UserId: orderReq.Userid,
			Amount: total,
			Products: orderReq.Items,
		}

		if err := kafka.ProducePaymentOrder(orderEvent); err != nil{
			log.Printf("Failed to produce payment order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully"})
	}


}

func calculateAmount(resp *cartpb.CartResponse) float64{
	var total float64
	for _, item := range resp.Items{
		total += float64(item.Quantity) * float64(item.Price)
	}
	return total
}