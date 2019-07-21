package main

import (
	"crypto/tls"
	"database/sql"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cooldryplace/cart"

	"github.com/cooldryplace/proto"

	"contrib.go.opencensus.io/exporter/prometheus"
	_ "github.com/lib/pq"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
)

const (
	defaultHTTPBind = ":8000"
	defaultGRPCBind = ":9000"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var (
		certFile = strings.TrimSpace(os.Getenv("TLS_CERT"))
		keyFile  = strings.TrimSpace(os.Getenv("TLS_CERT_KEY"))
	)

	if certFile == "" {
		log.Fatalf("TLS_CERT env var not set")
	}
	if keyFile == "" {
		log.Fatalf("TLS_CERT_KEY env var not set")
	}

	httpBind := strings.TrimSpace(os.Getenv("HTTP_BIND"))
	if httpBind == "" {
		log.Printf("Using default HTTP bind %q", defaultHTTPBind)
		httpBind = defaultHTTPBind
	}

	grpcBind := strings.TrimSpace(os.Getenv("GRPC_BIND"))
	if grpcBind == "" {
		log.Printf("Using default grpc bind %q", defaultGRPCBind)
		grpcBind = defaultGRPCBind
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

	dbConnStr := strings.TrimSpace(os.Getenv("DB_URL"))
	if dbConnStr == "" {
		log.Fatalf("DB_URL not set")
	}

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Failed to configure DB connection: %s", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to establish DB connection: %s", err)
	}

	grpcServer := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))

	proto.RegisterCartsServer(
		grpcServer,
		cart.NewServer(cart.New(cart.NewStorage(db))),
	)

	var (
		errChan    = make(chan error, 2)
		signalChan = make(chan os.Signal, 1)
	)

	go func() {
		log.Printf("gRPC listening on %s", grpcBind)
		errChan <- grpcServer.Serve(lis)
	}()

	pe, err := prometheus.NewExporter(prometheus.Options{Namespace: "cart"})
	if err != nil {
		log.Fatalf("failed to init Prometheus exporter: %s", err)
	}

	view.RegisterExporter(pe)
	view.SetReportingPeriod(1 * time.Second)

	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		log.Fatalf("Failed to register default server views: %s", err)
	}
	if err := view.Register(ocgrpc.DefaultClientViews...); err != nil {
		log.Fatalf("Failed to register default client views: %s", err)
	}

	http.Handle("/metrics", pe)
	http.HandleFunc("/health", healthCheck)

	go func() {
		log.Printf("HTTP listening on %s", httpBind)
		errChan <- http.ListenAndServe(httpBind, nil)
	}()

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Println(err)
	case <-signalChan:
		log.Println("Interrupt received. Graceful shutdown.")
	}

	grpcServer.GracefulStop()
}
