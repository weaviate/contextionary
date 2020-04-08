package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"

	goswagger "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	pb "github.com/semi-technologies/contextionary/contextionary"
	"github.com/semi-technologies/weaviate/client"
	"github.com/semi-technologies/weaviate/client/graphql"
	"github.com/semi-technologies/weaviate/client/things"
	"github.com/semi-technologies/weaviate/entities/models"
	"google.golang.org/grpc"
)

type mainCategory struct {
	ID          strfmt.UUID
	Name        string
	Description string
	Vector      []float32
}

var weaviate *client.Weaviate
var c11y pb.ContextionaryClient
var mainCategories []mainCategory
var allItems []interface{}

// var exampleCorpus = "I have used both my serial ports with a modem and a serial printer, so I cannot use Appletalk.  Is there a Ethernet to Localtalk hardware that will let me use the Ethernet port on my Q700 as a Localtalk  port.  Until they come out with satellite dishes that sit on your window & give you internet access from your home, I won't at all be using that port.  Saurabh.  "

func init() {
	initMainCategories()
	transport := goswagger.New("localhost:8080", "/v1", []string{"http"})
	weaviate = client.New(transport, strfmt.Default)

	conn, err := grpc.Dial("localhost:9999", grpc.WithInsecure())
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't connect: %s", err)
		os.Exit(1)
	}

	c11y = pb.NewContextionaryClient(conn)

	allItems = initSourceItems()

}

func initSourceItems() []interface{} {
	query := `{
  Get {
    Things {
      Post(where: {operator: Equal, path: ["training"], valueBoolean: false}, limit: 10000) {
        OfMainCategory {
          ... on MainCategory {
            name
          }
        }
        content
				controlMainCategoryId
      }
    }
  }
}
`

	res, err := weaviate.Graphql.GraphqlPost(graphql.NewGraphqlPostParams().WithBody(&models.GraphQLQuery{
		Query: query,
	}), nil)

	if err != nil {
		log.Fatal(err)
	}

	list := res.Payload.Data["Get"].(map[string]interface{})["Things"].(map[string]interface{})["Post"].([]interface{})
	return list
}

func chooseRandomItem(list []interface{}) (string, string, string) {
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(list))
	elem := list[i].(map[string]interface{})
	corpus := elem["content"].(string)
	controlID := strfmt.UUID(elem["controlMainCategoryId"].(string))
	control := mainCategoryById(controlID).Name
	oldPrediction := elem["OfMainCategory"].([]interface{})[0].(map[string]interface{})["name"].(string)

	return corpus, oldPrediction, control
}

func chooseItemAt(list []interface{}, i int) (string, string, string) {
	elem := list[i].(map[string]interface{})
	corpus := elem["content"].(string)
	controlID := strfmt.UUID(elem["controlMainCategoryId"].(string))
	control := mainCategoryById(controlID).Name
	oldPrediction := elem["OfMainCategory"].([]interface{})[0].(map[string]interface{})["name"].(string)

	return corpus, oldPrediction, control
}

func mainCategoryById(id strfmt.UUID) mainCategory {
	for _, cat := range mainCategories {
		if cat.ID == id {
			return cat
		}

	}

	panic("main cat not found by id")
}

func enrichWithVectors(in []mainCategory) []mainCategory {
	for i, obj := range in {
		res, err := weaviate.Things.ThingsGet(things.NewThingsGetParams().WithID(obj.ID).WithMeta(ptBool(true)), nil)
		if err != nil {
			log.Fatal(err)
		}

		in[i].Vector = res.Payload.Meta.Vector
	}

	return in
}

func initMainCategories() {
	mainCategories = []mainCategory{
		mainCategory{
			ID:          "51dd8b95-9e80-4824-9229-21f40e9b4e85",
			Name:        "Computers",
			Description: "Anything related to computers and their operating systems. Includes Apple Macintosh and Microsoft Windows related software.",
		},
		mainCategory{
			ID:          "e546bbab-fcc2-4688-8803-b75b061cc349",
			Name:        "Recreation",
			Description: "Leisure time activies and sports for recreational purposes. Includes baseball and hockey, but also motorsports and car discussions.",
		},
		mainCategory{
			ID:          "c2d5a423-5e5a-41df-b7b0-fd7e159482a3",
			Name:        "Science",
			Description: "Scientific Studies and experiments by universities, scientists and medical doctors.",
		},
		mainCategory{
			ID:          "ed0bab28-1479-4970-ad4f-c07ee6502da8",
			Name:        "For Sale",
			Description: "A marketplace for items items to sell and buy",
		},
		mainCategory{
			ID:          "d0ddbcc2-964a-4211-9ddc-d5d366e0dc14",
			Name:        "Politics",
			Description: "Political discussions focused mainly on the United States, but also includes world-wide politics. Hot topics include the political situation in the middle east as well as gun control",
		},
		mainCategory{
			ID:          "74e76a5b-8b4c-46b5-9898-e6b569c18a00",
			Name:        "Religion",
			Description: "Discussion about religion and atheism, includes Christianity, Islam and Jewish religions. Contains debates about wheter a God exists",
		},
	}
}

type doc struct {
	corpus        string
	oldPrediction string
	control       string
	newPrediction string
}

func main() {
	// total := 25
	total := len(allItems)

	mainCategories = enrichWithVectors(mainCategories)
	docs := make([]doc, total)

	fmt.Printf("Analyzing items")
	tfidf := NewTfIdfCalculator(total)

	for i := 0; i < total; i++ {
		fmt.Printf(".")
		// corpus, oldPrediction, control := chooseRandomItem(allItems)
		corpus, oldPrediction, control := chooseItemAt(allItems, i)
		tfidf.AddDoc(corpus)
		docs[i] = doc{corpus: corpus, oldPrediction: oldPrediction, control: control}
	}
	fmt.Printf("\n\n")

	fmt.Printf("Calculating tf-idf")
	tfidf.Calculate()
	fmt.Printf("\n\n")

	fmt.Printf("Classifying items")
	for i := 0; i < total; i++ {
		fmt.Printf(".")
		doc := docs[i]
		c := NewClassifier(i, doc, tfidf)
		tfIdfTerms := tfidf.GetAllTerms(i)

		newPredictions := c.Run()

		// always use top percentile
		// this is subject to change
		docs[i].newPrediction = newPredictions[0].Prediction

		serveResults(i, total, doc.corpus, doc.oldPrediction, doc.control, c.ranked, newPredictions, tfIdfTerms)
		c = nil
	}
	fmt.Printf("\n\n")

	fmt.Printf("Serving on port 7070")
	calculateSuccessAndServe(docs)
	http.ListenAndServe(":7070", nil)
}

func calculateSuccessAndServe(docs []doc) {
	total := float32(len(docs))
	var previouslyCorrect uint
	var newlyCorrect uint

	for _, doc := range docs {
		if doc.oldPrediction == doc.control {
			previouslyCorrect++
		}

		if doc.newPrediction == doc.control {
			newlyCorrect++
		}
	}

	newSuccessRate := float32(newlyCorrect) / total
	previousSuccessRate := float32(previouslyCorrect) / total

	absoluteImprovement := newSuccessRate - previousSuccessRate
	relativeImprovement := newSuccessRate / previousSuccessRate

	serveSuccess(docs, total, newSuccessRate, previousSuccessRate, absoluteImprovement, relativeImprovement)
}

func rank(in []wordWithDistance) []wordWithDistance {
	i := 0
	filtered := make([]wordWithDistance, len(in))
	for _, w := range in {
		if w.Distance < -9 {
			continue
		}

		filtered[i] = w
		i++
	}
	filtered = filtered[:i]
	sort.Slice(filtered, func(a, b int) bool { return filtered[a].InformationGain > filtered[b].InformationGain })
	return filtered
}

type wordWithDistance struct {
	Word            string
	Distance        float32
	AverageDistance float32
	Prediction      string
	InformationGain float32
}

func extractVector(in *pb.Vector) []float32 {
	out := make([]float32, len(in.Entries))
	for i, elem := range in.Entries {
		out[i] = elem.Entry
	}

	return out
}

func ptBool(in bool) *bool {
	return &in
}

func NewSplitter() *Splitter {
	return &Splitter{}
}

type Splitter struct{}

func (s *Splitter) Split(corpus string) []string {
	return strings.FieldsFunc(corpus, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})
}

func cosineSim(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors have different dimensions")
	}

	var (
		sumProduct float64
		sumASquare float64
		sumBSquare float64
	)

	for i := range a {
		sumProduct += float64(a[i] * b[i])
		sumASquare += float64(a[i] * a[i])
		sumBSquare += float64(b[i] * b[i])
	}

	return float32(sumProduct / (math.Sqrt(sumASquare) * math.Sqrt(sumBSquare))), nil
}

func cosineDist(a, b []float32) (float32, error) {
	sim, err := cosineSim(a, b)
	if err != nil {
		return 0, err
	}

	return 1 - sim, nil
}
