package main

import (
	"context"
	"fmt"
	"os"

	pb "github.com/creativesoftwarefdn/contextionary/contextionary"
	grpc "google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:9999", grpc.WithInsecure())
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't connect: %s", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := pb.NewContextionaryClient(conn)

	ctx := context.Background()
	res, err := client.GetHello(ctx, &pb.Void{})
	fmt.Printf("%#v %#v", res, err)
}
