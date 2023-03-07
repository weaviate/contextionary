package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/weaviate/contextionary/contextionary"
	"google.golang.org/grpc"
)

var expectedDimensions int

func init() {

	d, err := strconv.Atoi(os.Getenv("DIMENSIONS"))
	if err != nil {
		panic(err)
	}

	expectedDimensions = d
}

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

		t.Run("corpi to vector", func(t *testing.T) {
			t.Run("only stopwords", func(t *testing.T) {
				corpi := []string{"of", "the of"}
				_, err := client.VectorForCorpi(context.Background(), &pb.Corpi{Corpi: corpi})
				assert.NotNil(t, err)
			})

			t.Run("only stopwords", func(t *testing.T) {
				corpi := []string{"car", "car of brand mercedes", "color blue"}
				res, err := client.VectorForCorpi(context.Background(), &pb.Corpi{Corpi: corpi})
				assert.Nil(t, err)
				// TODO: also upgrade minimal one to 600 vectors
				assert.Len(t, res.Entries, 300)
			})

			t.Run("two corpi with and without splitting characters should lead to the same vector", func(t *testing.T) {
				corpi1 := []string{"car", "car of brand mercedes", "color blue"}
				corpi2 := []string{"car,", "car#of,,,,brand<mercedes", "color!!blue"}
				res1, err := client.VectorForCorpi(context.Background(), &pb.Corpi{Corpi: corpi1})
				assert.Nil(t, err)
				assert.Len(t, res1.Entries, 300)

				res2, err := client.VectorForCorpi(context.Background(), &pb.Corpi{Corpi: corpi2})
				assert.Nil(t, err)
				assert.Len(t, res2.Entries, 300)

				assert.Equal(t, res1.Entries, res2.Entries)
			})
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
			words := []string{"the", "a"}

			for _, word := range words {
				t.Run(word, func(t *testing.T) {
					res, err := client.IsWordStopword(context.Background(), &pb.Word{Word: word})
					require.Nil(t, err)
					assert.Equal(t, true, res.Stopword)
				})
			}
		})

		t.Run("corpi to vector", func(t *testing.T) {
			t.Run("only stopwords", func(t *testing.T) {
				corpi := []string{"a", "the a"}
				_, err := client.VectorForCorpi(context.Background(), &pb.Corpi{Corpi: corpi})
				assert.NotNil(t, err)
			})

			t.Run("not only stopwords", func(t *testing.T) {
				corpi := []string{"car", "car of brand mercedes", "color blue"}
				res, err := client.VectorForCorpi(context.Background(), &pb.Corpi{Corpi: corpi})
				require.Nil(t, err)
				fmt.Println(expectedDimensions)
				fmt.Println(res)
				assert.Len(t, res.Entries, expectedDimensions)
			})

			t.Run("two corpi with and without splitting characters should lead to the same vector", func(t *testing.T) {
				corpi1 := []string{"car", "car of brand mercedes", "color blue"}
				corpi2 := []string{"car,", "car#of,,,,brand<mercedes", "color!!blue"}
				res1, err := client.VectorForCorpi(context.Background(), &pb.Corpi{Corpi: corpi1})
				require.Nil(t, err)
				assert.Len(t, res1.Entries, expectedDimensions)

				res2, err := client.VectorForCorpi(context.Background(), &pb.Corpi{Corpi: corpi2})
				require.Nil(t, err)
				assert.Len(t, res2.Entries, expectedDimensions)

				assert.Equal(t, res1.Entries, res2.Entries)
			})
		})
	})
}
