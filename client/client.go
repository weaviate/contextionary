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

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "no command provided, try 'word-present'\n")
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "word-present":
		wordPresent(client, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command '%s'\n", cmd)
		os.Exit(1)
	}
}

func wordPresent(client pb.ContextionaryClient, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "need at least one other argument: the word you want to check\n")
		os.Exit(1)
	}

	ctx := context.Background()

	for _, word := range args {
		res, err := client.IsWordPresent(ctx, &pb.Word{Word: word})
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: couldn't get word: %s", err)
			os.Exit(1)
		}
		if res.Present {
			fmt.Printf("word '%s' is present in the contextionary\n", word)
		} else {
			fmt.Printf("word '%s' is NOT present in the contextionary\n", word)
		}
	}
}
