package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cooldryplace/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	grpcCAEnv   = "GRPC_CA"
	defaultBind = "localhost:9000"
)

var userID = time.Now().Unix() * -1

func printCart(ctx context.Context, client proto.CartsClient, cartID int64) {
	resp, err := client.GetCart(ctx, &proto.CartRequest{Id: cartID})
	if err != nil {
		log.Fatalf("Failed to get created cart: %s", err)
	}

	cart := resp.Cart

	fmt.Println("░░░░░░░░░░░░░░░░░░░░░░░░░░░")
	fmt.Println("╔══════════════════════════╗")
	fmt.Printf("║\tCART: %d\n", cart.Id)
	fmt.Println("║══════════════════════════║")
	fmt.Printf("║ UserId: \t%d\n", cart.UserId)
	fmt.Println("║──────────────────────────║")

	fmt.Println("║ # \tProduct\tQuantity")
	for i, item := range cart.Items {
		fmt.Printf("║ %d: \t%d\t%d\n", i, item.ProductId, item.Quantity)
	}

	fmt.Println("╚══════════════════════════╝")
	fmt.Println()
}

func main() {
	bind := flag.String("bind", defaultBind, "Carts service bind")

	flag.Parse()

	caFile := strings.TrimSpace(os.Getenv(grpcCAEnv))
	if caFile == "" {
		log.Fatalf("%s env var not set", grpcCAEnv)
	}

	creds, err := credentials.NewClientTLSFromFile(caFile, "")
	if err != nil {
		log.Fatalf("Failed to create TLS credentials %v", err)
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *bind, opts...)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	client := proto.NewCartsClient(conn)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	resp, err := client.CreateCart(ctx, &proto.CartCreateRequest{UserId: userID})
	if err != nil {
		log.Fatalf("Failed to create a Cart: %s", err)
	}

	cartID := resp.Cart.Id

	printCart(ctx, client, cartID)

	if _, err := client.AddProduct(ctx, &proto.AddProductRequest{CartId: cartID, ProductId: 42, Quantity: 10}); err != nil {
		log.Fatalf("Failed to add a Product: %s", err)
	}

	printCart(ctx, client, cartID)

	if _, err := client.AddProduct(ctx, &proto.AddProductRequest{CartId: cartID, ProductId: 95, Quantity: 1}); err != nil {
		log.Fatalf("Failed to add a Product: %s", err)
	}

	printCart(ctx, client, cartID)

	if _, err := client.AddProduct(ctx, &proto.AddProductRequest{CartId: cartID, ProductId: 95, Quantity: 1}); err != nil {
		log.Fatalf("Failed to add a Product: %s", err)
	}

	printCart(ctx, client, cartID)

	if _, err := client.DelProduct(ctx, &proto.DelProductRequest{CartId: cartID, ProductId: 42}); err != nil {
		log.Fatalf("Failed to delete a Product: %s", err)
	}

	printCart(ctx, client, cartID)

	if _, err := client.EmptyCart(ctx, &proto.EmptyCartRequest{CartId: cartID}); err != nil {
		log.Fatalf("Failed to empty a Cart: %s", err)
	}

	printCart(ctx, client, cartID)

	if _, err := client.DeleteCart(ctx, &proto.CartDeleteRequest{Id: cartID}); err != nil {
		log.Fatalf("Failed to delete a Cart: %s", err)
	}

	printCart(ctx, client, cartID)

}
