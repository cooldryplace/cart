package cart

import (
	"context"
	"crypto/tls"
	"database/sql"
	"flag"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cooldryplace/proto"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var cartsClient proto.CartsClient

const (
	grpcBind  = "localhost:9001"
	grpcCAEnv = "GRPC_CA"
)

func client(bind, caFile string) proto.CartsClient {
	creds, err := credentials.NewClientTLSFromFile(caFile, "")
	if err != nil {
		log.Fatalf("Failed to create TLS credentials %v", err)
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, bind, opts...)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	return proto.NewCartsClient(conn)
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		return
	}

	var (
		caFile   = strings.TrimSpace(os.Getenv("GRPC_CA"))
		certFile = strings.TrimSpace(os.Getenv("TLS_CERT"))
		keyFile  = strings.TrimSpace(os.Getenv("TLS_CERT_KEY"))
	)

	if caFile == "" {
		log.Fatal("GRPC_CA env var not set")
	}
	if certFile == "" {
		log.Fatal("TLS_CERT env var not set")
	}
	if keyFile == "" {
		log.Fatal("TLS_CERT_KEY env var not set")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	lis, err := tls.Listen("tcp", grpcBind, &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	})
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()

	dbConnStr := strings.TrimSpace(os.Getenv("TEST_DB_URL"))
	if dbConnStr == "" {
		log.Fatalf("TEST_DB_URL not set")
	}

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Failed to configure DB connection: %s", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to establish DB connection: %s", err)
	}

	proto.RegisterCartsServer(
		grpcServer,
		NewServer(New(NewStorage(db))),
	)

	go grpcServer.Serve(lis)
	cartsClient = client(grpcBind, caFile)

	os.Exit(m.Run())
}

func cartByID(ctx context.Context, t *testing.T, cartID int64) *proto.Cart {
	t.Helper()

	resp, err := cartsClient.GetCart(ctx, &proto.CartRequest{Id: cartID})
	if err != nil {
		t.Fatalf("Failed to get created cart: %s", err)
	}

	return resp.Cart
}

func TestUnknownCartReturnsNotFound(t *testing.T) {
	var (
		ctx                 = context.Background()
		unknownCartID int64 = 4444444
	)

	_, err := cartsClient.GetCart(ctx, &proto.CartRequest{Id: unknownCartID})
	if err == nil {
		t.Fatal("Expected to get error")
	}

	actual := status.Code(err)
	expected := codes.NotFound

	if actual != expected {
		t.Errorf("Got status: %v, expected: %v", actual, expected)
	}
}

func TestCartCreate(t *testing.T) {
	var (
		ctx          = context.Background()
		userID int64 = 314
	)

	resp, err := cartsClient.CreateCart(ctx, &proto.CartCreateRequest{UserId: userID})
	if err != nil {
		t.Fatalf("Failed to create a Cart: %s", err)
	}

	cartID := resp.Cart.Id

	actual := cartByID(ctx, t, cartID)
	if actual.UserId != userID {
		t.Errorf("Got user ID: %d, expected: %d", actual.UserId, userID)
	}

	if len(actual.Items) != 0 {
		t.Error("Expected new Cart to be empty")
	}
}
