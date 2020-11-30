package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danikarik/product-storage/pkg/api"
)

var (
	addr    = flag.String("http.addr", ":8080", "address for listening")
	timeout = flag.Duration("fetch.timeout", 30*time.Second, "address for listening")
)

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var (
		srv  = api.NewServer(&api.Options{Timeout: *timeout})
		exit = make(chan error, 1)
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
