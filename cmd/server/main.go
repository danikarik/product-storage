package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danikarik/product-storage/pkg/api"
	"github.com/danikarik/product-storage/pkg/repo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	addr    = flag.String("http.addr", ":50051", "address to listen")
	dbhost  = flag.String("db.host", "mongodb://localhost:27017", "mongo host address")
	timeout = flag.Duration("fetch.timeout", 30*time.Second, "address for listening")
)

func main() {
	ctx := context.Background()

	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("net listener: %v", err)
	}

	conn, err := mongo.NewClient(options.Client().ApplyURI(*dbhost))
	if err != nil {
		log.Fatalf("mongo client: %v", err)
	}

	if err := conn.Connect(ctx); err != nil {
		log.Fatalf("mongo connection: %v", err)
	}
	defer conn.Disconnect(ctx)

	var (
		exit = make(chan error, 1)
		srv  = api.NewServer(
			repo.NewMongoRepo("productstore", conn),
			&api.Options{Timeout: *timeout},
		)
	)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		exit <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		log.Println("start listening on: " + *addr)
		exit <- srv.Serve(listener)
	}()

	<-exit

	fmt.Println("")
	log.Println("shutting down server ...")

	if err := listener.Close(); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}
}
