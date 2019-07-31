package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	pb "github.com/semi-technologies/contextionary/contextionary"
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
	case "meta", "version":
		meta(client, args[1:])
	case "word-present":
		wordPresent(client, args[1:])
	case "word-stopword":
		wordStopword(client, args[1:])
	case "search":
		search(client, args[1:])
	case "similar-words":
		similarWords(client, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command '%s'\n", cmd)
		os.Exit(1)
	}
}
func meta(client pb.ContextionaryClient, args []string) {
	ctx := context.Background()

	res, err := client.Meta(ctx, &pb.MetaParams{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: couldn't display meta: %s", err)
		os.Exit(1)
	}

	fmt.Printf("%#v\n", res)
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

func similarWords(client pb.ContextionaryClient, args []string) {
	var word string
	var certainty float32

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "need at least one other argument: the word you want to find similarities to\n")
		os.Exit(1)
	}
	word = args[0]

	if len(args) == 1 {
		fmt.Fprintf(os.Stderr, "need at least one other argument: the minimum required certainty\n")
		os.Exit(1)
	}

	c, err := strconv.ParseFloat(args[1], 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldnt parse certainty: %v\n", err)
		os.Exit(1)
	}
	certainty = float32(c)

	res, err := client.SafeGetSimilarWordsWithCertainty(context.Background(), &pb.SimilarWordsParams{
		Certainty: certainty,
		Word:      word,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: couldn't get similar words: %s", err)
		os.Exit(1)
	}

	for _, word := range res.Words {
		fmt.Printf("🥳  %s\n", word.Word)
	}
}

func wordStopword(client pb.ContextionaryClient, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "need at least one other argument: the word you want to check\n")
		os.Exit(1)
	}

	ctx := context.Background()

	for _, word := range args {
		res, err := client.IsWordStopword(ctx, &pb.Word{Word: word})
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: couldn't get word: %s", err)
			os.Exit(1)
		}
		if res.Stopword {
			fmt.Printf("word '%s' is a stopword\n", word)
		} else {
			fmt.Printf("word '%s' is not a stopword\n", word)
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
