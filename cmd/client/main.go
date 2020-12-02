package main

import (
	"context"
	"flag"
	"log"

	pb "github.com/danikarik/product-storage/pkg/store"
	"google.golang.org/grpc"
)

var (
	addr  = flag.String("http.addr", ":50051", "server host address")
	fetch = flag.String("fetch.url", "https://csv-samples.s3.amazonaws.com/dummy.csv", "data url to be fetched")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	conn, err := grpc.Dial(*addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	defer conn.Close()

	c := pb.NewStoreClient(conn)

	if _, err = c.Fetch(ctx, &pb.FetchRequest{Url: *fetch}); err != nil {
		log.Fatalf("fetch failed: %v", err)
	}

	resp, err := c.List(ctx, &pb.ListRequest{
		Paging:  &pb.Paging{Limit: 5},
		Sorting: &pb.Sorting{Field: pb.Field_NAME, Direction: pb.Direction_DESC},
	})

	log.Println(resp.LastId)

	for _, p := range resp.Products {
		log.Printf("%s | %f | %d | %s \n", p.Name, p.Price, p.NumOfChanges, p.LastUpdate)
	}
}
