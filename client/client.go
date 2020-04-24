package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	pb "github.com/semi-technologies/contextionary/contextionary"
	grpc "google.golang.org/grpc"
)

func help() {
	fmt.Println("the following commands are supported:")
	fmt.Printf("\n")
	fmt.Printf("\t%-15s%s\n", "meta", "Display meta info, such as versions")
	fmt.Printf("\t               %s\n", "Usage: client meta")
	fmt.Printf("\n")
	fmt.Printf("\t%-15s%s\n", "word-present", "Check if the word is present in the db or as an extension")
	fmt.Printf("\t               %s\n", "Usage: client word-present word")
	fmt.Printf("\n")
	fmt.Printf("\t%-15s%s\n", "word-stopword", "Check if the word is considered a stopword")
	fmt.Printf("\t               %s\n", "Usage: client word-stopword word")
	fmt.Printf("\n")
	fmt.Printf("\t%-15s%s\n", "search", "Search for word or property")
	fmt.Printf("\t               %s\n", "For usage run client search and see instructions from there")
	fmt.Printf("\n")
	fmt.Printf("\t%-15s%s\n", "similar-words", "Search for similar words within the specified certainty")
	fmt.Printf("\t               %s\n", "Usage: client similar-words word certainty")
	fmt.Printf("\n")
	fmt.Printf("\t%-15s%s\n", "extend", "Extend the contextionary with custom concepts")
	fmt.Printf("\t               %s\n", "Usage: client extend newconcept \"definition of the new concept\"")
	fmt.Printf("\n")
	fmt.Printf("\t%-15s%s\n", "vectorize", "Vectorize any string")
	fmt.Printf("\t               %s\n", "Usage: client vectorize \"input string to vectorize\"")
	fmt.Printf("\t%-15s%s\n", "multi-vector-for-word", "Vectorize multiple strings")
	fmt.Printf("\t               %s\n", "Usage: client multi-vector-for-word \"word1 word2 word3 ... wordN\"")
}

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
	case "help":
		help()
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
	case "extend":
		extend(client, args[1:])
	case "vectorize":
		vectorize(client, args[1:])
	case "multi-vector-for-word":
		multiVecForWord(client, args[1:])

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
func extend(client pb.ContextionaryClient, args []string) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "need two arguments, the concept to add/extend and its definition\n")
		os.Exit(1)
	}
	concept := args[0]
	definition := args[1]

	_, err := client.AddExtension(context.Background(), &pb.ExtensionInput{
		Concept:    concept,
		Definition: definition,
		Weight:     1,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s", err)
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stdout, "Success!")
		os.Exit(0)
	}
}

func vectorize(client pb.ContextionaryClient, args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "need one argument: the input string to vectorize")
		os.Exit(1)
	}
	input := args[0]

	res, err := client.VectorForCorpi(context.Background(), &pb.Corpi{
		Corpi: []string{input},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s", err)
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stdout, "Success: %v", res.Entries)
		os.Exit(0)
	}
}

func multiVecForWord(client pb.ContextionaryClient, args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "need at least one argument: the input word to vectorize")
		os.Exit(1)
	}

	words := make([]*pb.Word, len(args))
	for i, word := range args {
		words[i] = &pb.Word{Word: word}
	}

	res, err := client.MultiVectorForWord(context.Background(), &pb.WordList{
		Words: words,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s", err)
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stdout, "Success: %v", res.Vectors)
		os.Exit(0)
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
