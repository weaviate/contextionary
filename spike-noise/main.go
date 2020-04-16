package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
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

type target struct {
	ID     strfmt.UUID
	Name   string
	Vector []float32
}

const (
	SourceClassName                 = "Post"
	TargetClassName                 = "MainCategory"
	TargetPropertyName              = "name"
	ClassifyProperty                = "ofMainCategory"
	ControlProperty                 = "controlMainCategory"
	BasedOnProperty                 = "content"
	WhereFilterToFindUnlabelledData = "{operator: Equal, path: [\"training\"], valueBoolean: false}"
)

var weaviate *client.Weaviate
var c11y pb.ContextionaryClient
var targets []target
var allItems []interface{}

func init() {
	transport := goswagger.New("localhost:8080", "/v1", []string{"http"})
	weaviate = client.New(transport, strfmt.Default)

	conn, err := grpc.Dial("localhost:9999", grpc.WithInsecure())
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't connect: %s", err)
		os.Exit(1)
	}

	c11y = pb.NewContextionaryClient(conn)

	targets = initTargets()
	allItems = initSourceItems()

}

func uppercaseFirstLetter(in string) string {
	return fmt.Sprintf("%s%s", strings.ToUpper(string(in[:1])), string(in[1:]))
}

func initSourceItems() []interface{} {
	query := fmt.Sprintf(`{
  Get {
    Things {
      %s (where: %s, limit: 10000) {
        %s {
          ... on %s {
            %s
          }
        }
        %s
        %s {
          ... on %s {
            %s
          }
        }
      }
    }
  }
}
`, SourceClassName, WhereFilterToFindUnlabelledData, uppercaseFirstLetter(ClassifyProperty), TargetClassName, TargetPropertyName, BasedOnProperty,
		uppercaseFirstLetter(ControlProperty), TargetClassName, TargetPropertyName)

	res, err := weaviate.Graphql.GraphqlPost(graphql.NewGraphqlPostParams().WithBody(&models.GraphQLQuery{
		Query: query,
	}), nil)

	if err != nil {
		log.Fatal(err)
	}

	if err := res.Payload.Errors; err != nil {
		log.Fatal(err[0])
	}

	list := res.Payload.Data["Get"].(map[string]interface{})["Things"].(map[string]interface{})[SourceClassName].([]interface{})
	return list
}

func chooseItemAt(list []interface{}, i int) (string, string, string) {
	elem := list[i].(map[string]interface{})
	corpus := elem[BasedOnProperty].(string)
	control := elem[uppercaseFirstLetter(ControlProperty)].([]interface{})[0].(map[string]interface{})[TargetPropertyName].(string)
	oldPrediction := elem[uppercaseFirstLetter(ClassifyProperty)].([]interface{})[0].(map[string]interface{})[TargetPropertyName].(string)

	return corpus, oldPrediction, control
}

func enrichWithVectors(in []target) []target {
	for i, obj := range in {
		res, err := weaviate.Things.ThingsGet(things.NewThingsGetParams().WithID(obj.ID).WithMeta(ptBool(true)), nil)
		if err != nil {
			log.Fatal(err)
		}

		in[i].Vector = res.Payload.Meta.Vector
	}

	return in
}

func initTargets() []target {
	query := fmt.Sprintf(`{
  Get {
    Things {
      %s {
			  uuid
				%s
      }
    }
  }
}
`, TargetClassName, TargetPropertyName)

	res, err := weaviate.Graphql.GraphqlPost(graphql.NewGraphqlPostParams().WithBody(&models.GraphQLQuery{
		Query: query,
	}), nil)

	if err != nil {
		log.Fatal(err)
	}

	if err := res.Payload.Errors; err != nil {
		log.Fatal(err[0])
	}

	list := res.Payload.Data["Get"].(map[string]interface{})["Things"].(map[string]interface{})[TargetClassName].([]interface{})

	targets := make([]target, len(list))
	for i, obj := range list {
		targets[i].ID = strfmt.UUID(obj.(map[string]interface{})["uuid"].(string))
		targets[i].Name = obj.(map[string]interface{})[TargetPropertyName].(string)
	}

	return enrichWithVectors(targets)
}

type doc struct {
	corpus        string
	oldPrediction string
	control       string
	newPrediction string
}

func main() {
	total := len(allItems)
	docs := make([]doc, total)

	fmt.Printf("Analyzing items")
	tfidf := NewTfIdfCalculator(total)

	for i := 0; i < total; i++ {
		fmt.Printf(".")
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

	byLength := calculateSuccessByLength(docs)
	serveSuccess(docs, total, newSuccessRate, previousSuccessRate, absoluteImprovement, relativeImprovement, byLength)
}

type bucket struct {
	Words    int
	Success  uint
	Total    uint
	Ratio    float64
	Elements []bucketElement
}

type bucketElement struct {
	Index   int
	Success bool
}

func increaseSuccess(buckets map[int]*bucket, index int) {
	if _, ok := buckets[index]; !ok {
		buckets[index] = &bucket{
			Words: index,
		}
	}

	buckets[index].Success++
}

func increaseTotal(buckets map[int]*bucket, index int, docIndex int, success bool) {
	if _, ok := buckets[index]; !ok {
		buckets[index] = &bucket{
			Words: index,
		}
	}

	buckets[index].Total++
	buckets[index].Elements = append(buckets[index].Elements, bucketElement{Index: docIndex, Success: success})
}

func calculateSuccessByLength(docs []doc) []*bucket {
	buckets := map[int]*bucket{}
	for i, doc := range docs {
		words := len(NewSplitter().Split(doc.corpus))
		success := doc.newPrediction == doc.control
		if success {
			switch true {
			case words < 5:
				increaseSuccess(buckets, 5)
			case words < 10:
				increaseSuccess(buckets, 10)
			case words < 20:
				increaseSuccess(buckets, 20)
			case words < 40:
				increaseSuccess(buckets, 40)
			case words < 80:
				increaseSuccess(buckets, 80)
			case words < 160:
				increaseSuccess(buckets, 160)
			case words < 320:
				increaseSuccess(buckets, 320)
			case words < 640:
				increaseSuccess(buckets, 640)
			case words < 1280:
				increaseSuccess(buckets, 1280)
			case words < 2560:
				increaseSuccess(buckets, 2560)
			default:
				increaseSuccess(buckets, 2561)
			}
		}

		switch true {
		case words < 5:
			increaseTotal(buckets, 5, i, success)
		case words < 10:
			increaseTotal(buckets, 10, i, success)
		case words < 20:
			increaseTotal(buckets, 20, i, success)
		case words < 40:
			increaseTotal(buckets, 40, i, success)
		case words < 80:
			increaseTotal(buckets, 80, i, success)
		case words < 160:
			increaseTotal(buckets, 160, i, success)
		case words < 320:
			increaseTotal(buckets, 320, i, success)
		case words < 640:
			increaseTotal(buckets, 640, i, success)
		case words < 1280:
			increaseTotal(buckets, 1280, i, success)
		case words < 2560:
			increaseTotal(buckets, 2560, i, success)
		default:
			increaseTotal(buckets, 2561, i, success)
		}
	}

	out := make([]*bucket, len(buckets))
	i := 0
	for _, value := range buckets {
		value.Ratio = float64(value.Success) / float64(value.Total)
		out[i] = value
		i++
	}

	sort.Slice(out, func(a, b int) bool { return out[a].Words < out[b].Words })

	return out
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
