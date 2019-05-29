package main

import (
	"context"
	"testing"

	pb "github.com/semi-technologies/contextionary/contextionary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func Test_Contextionary_Journey(t *testing.T) {
	// minimal
	connMinimal, err := grpc.Dial("minimal:9999", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("couldn't connect to minimal c11y: %s", err)
	}
	defer connMinimal.Close()

	connFull, err := grpc.Dial("full:9999", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("couldn't connect to minimal c11y: %s", err)
	}
	defer connFull.Close()

	clientMinimal := pb.NewContextionaryClient(connMinimal)
	clientFull := pb.NewContextionaryClient(connFull)

	t.Run("the minimal contextionary", func(t *testing.T) {
		client := clientMinimal

		t.Run("testing words present", func(t *testing.T) {
			words := []string{"car", "engine", "automobile", "name"}

			for _, word := range words {
				t.Run(word, func(t *testing.T) {
					res, err := client.IsWordPresent(context.Background(), &pb.Word{Word: word})
					require.Nil(t, err)
					assert.Equal(t, true, res.Present)
				})
			}
		})

		t.Run("testing stopwords", func(t *testing.T) {
			words := []string{"of", "the"}

			for _, word := range words {
				t.Run(word, func(t *testing.T) {
					res, err := client.IsWordStopword(context.Background(), &pb.Word{Word: word})
					require.Nil(t, err)
					assert.Equal(t, true, res.Stopword)
				})
			}
		})
	})

	t.Run("the full contextionary", func(t *testing.T) {
		client := clientFull

		t.Run("testing words present", func(t *testing.T) {
			words := []string{"car", "engine", "automobile", "influenza", "brexit", "condenser", "name"}

			for _, word := range words {
				t.Run(word, func(t *testing.T) {
					res, err := client.IsWordPresent(context.Background(), &pb.Word{Word: word})
					require.Nil(t, err)
					assert.Equal(t, true, res.Present)
				})
			}
		})

		t.Run("testing stopwords", func(t *testing.T) {
			words := []string{"of", "the", "a", "ah", "almost", "always", "also", "as"}

			for _, word := range words {
				t.Run(word, func(t *testing.T) {
					res, err := client.IsWordStopword(context.Background(), &pb.Word{Word: word})
					require.Nil(t, err)
					assert.Equal(t, true, res.Stopword)
				})
			}
		})
	})
}
