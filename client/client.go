package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

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
	case "search":
		search(client, args[1:])
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

func search(client pb.ContextionaryClient, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "need at least one other argument: either 'class' or 'property' \n")
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "class":
		searchClass(client, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command '%s'\n", cmd)
		os.Exit(1)
	}
}

func searchClass(client pb.ContextionaryClient, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "need at least one other argument the search term\n")
		os.Exit(1)
	}

	if len(args) == 1 {
		fmt.Fprintf(os.Stderr, "need at least one other argument the desired certainty\n")
		os.Exit(1)
	}

	searchTerm := args[0]
	certainty, err := strconv.ParseFloat(args[1], 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot parse certainty '%s'\n", args[1])
		os.Exit(1)
	}

	params := &pb.SchemaSearchParams{
		Certainty: float32(certainty),
		Name:      searchTerm,
		Kind:      pb.Kind_THING,
	}

	ctx := context.Background()
	res, err := client.SchemaSearch(ctx, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "schema search failed: %s", err)
		os.Exit(1)
	}

	if len(res.Results) == 0 {
		fmt.Println("😵 nothing found")
	}

	for _, class := range res.Results {
		fmt.Printf("🥳  %s (Certainty: %f)\n", class.Name, class.Certainty)
	}
}
