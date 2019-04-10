package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/james-bowman/nlp"
	"github.com/james-bowman/nlp/measures/pairwise"
	"github.com/turnage/graw/reddit"
	"gonum.org/v1/gonum/mat"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
)

type requestPayload struct {
	CommentToPost string `json:"content"`
	Path          string `json:"path"`
}

type redditComment struct {
	Id      string `json:"id"`
	Comment string `json:"comment"`
}

type similarComment struct {
	Ranking float64       `json:"similarity"`
	Comment redditComment `json:"comment"`
}

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

func HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.Headers["content-type"] != "application/json" {
		return clientError(http.StatusNotAcceptable)
	}

	var reqBody requestPayload
	err := json.Unmarshal([]byte(req.Body), &reqBody)
	if err != nil {
		return clientError(http.StatusUnprocessableEntity)
	}

	dir, err := os.Getwd()
	if err != nil {
		return serverError(err)
	}

	bot, err := reddit.NewBotFromAgentFile(dir+"/"+os.Getenv("AGENTFILE"), 0)
	if err != nil {
		return serverError(err)
	}

	post, err := bot.Thread(reqBody.Path)
	if err != nil {
		return serverError(err)
	}

	similarComments, err := getSimilarComments(reqBody.CommentToPost, post.Replies)

	if err != nil {
		return serverError(err)
	}

	js, err := json.Marshal(similarComments)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers: map[string]string{
			"Access-Control-Allow-Origin": os.Getenv("CORS_ORIGIN"),
		},
	}, nil
}

func getSimilarComments(query string, corpus []*reddit.Comment) ([]similarComment, error) {
	stopWords := []string{"a", "about", "above", "above", "across", "after", "afterwards", "again", "against", "all", "almost", "alone", "along", "already", "also", "although", "always", "am", "among", "amongst", "amoungst", "amount", "an", "and", "another", "any", "anyhow", "anyone", "anything", "anyway", "anywhere", "are", "around", "as", "at", "back", "be", "became", "because", "become", "becomes", "becoming", "been", "before", "beforehand", "behind", "being", "below", "beside", "besides", "between", "beyond", "bill", "both", "bottom", "but", "by", "call", "can", "cannot", "cant", "co", "con", "could", "couldnt", "cry", "de", "describe", "detail", "do", "done", "down", "due", "during", "each", "eg", "eight", "either", "eleven", "else", "elsewhere", "empty", "enough", "etc", "even", "ever", "every", "everyone", "everything", "everywhere", "except", "few", "fifteen", "fify", "fill", "find", "fire", "first", "five", "for", "former", "formerly", "forty", "found", "four", "from", "front", "full", "further", "get", "give", "go", "had", "has", "hasnt", "have", "he", "hence", "her", "here", "hereafter", "hereby", "herein", "hereupon", "hers", "herself", "him", "himself", "his", "how", "however", "hundred", "ie", "if", "in", "inc", "indeed", "interest", "into", "is", "it", "its", "itself", "keep", "last", "latter", "latterly", "least", "less", "ltd", "made", "many", "may", "me", "meanwhile", "might", "mill", "mine", "more", "moreover", "most", "mostly", "move", "much", "must", "my", "myself", "name", "namely", "neither", "never", "nevertheless", "next", "nine", "no", "nobody", "none", "noone", "nor", "not", "nothing", "now", "nowhere", "of", "off", "often", "on", "once", "one", "only", "onto", "or", "other", "others", "otherwise", "our", "ours", "ourselves", "out", "over", "own", "part", "per", "perhaps", "please", "put", "rather", "re", "same", "see", "seem", "seemed", "seeming", "seems", "serious", "several", "she", "should", "show", "side", "since", "sincere", "six", "sixty", "so", "some", "somehow", "someone", "something", "sometime", "sometimes", "somewhere", "still", "such", "system", "take", "ten", "than", "that", "the", "their", "them", "themselves", "then", "thence", "there", "thereafter", "thereby", "therefore", "therein", "thereupon", "these", "they", "thickv", "thin", "third", "this", "those", "though", "three", "through", "throughout", "thru", "thus", "to", "together", "too", "top", "toward", "towards", "twelve", "twenty", "two", "un", "under", "until", "up", "upon", "us", "very", "via", "was", "we", "well", "were", "what", "whatever", "when", "whence", "whenever", "where", "whereafter", "whereas", "whereby", "wherein", "whereupon", "wherever", "whether", "which", "while", "whither", "who", "whoever", "whole", "whom", "whose", "why", "will", "with", "within", "without", "would", "yet", "you", "your", "yours", "yourself", "yourselves"}

	vectoriser := nlp.NewCountVectoriser(stopWords...)
	transformer := nlp.NewTfidfTransformer()

	reducer := nlp.NewTruncatedSVD(4)

	lsiPipeline := nlp.NewPipeline(vectoriser, transformer, reducer)

	var comments []string
	for _, commentInfo := range corpus {
		comments = append(comments, commentInfo.Body)
	}

	lsi, err := lsiPipeline.FitTransform(comments...)
	if err != nil {
		return nil, fmt.Errorf("failed to process documents because %v", err)
	}

	queryVector, err := lsiPipeline.Transform(query)
	if err != nil {
		return nil, fmt.Errorf("failed to process documents because %v", err)
	}

	_, docs := lsi.Dims()
	matches := make([]similarComment, 0)
	ch := make(chan similarComment)
	var wg sync.WaitGroup
	wg.Add(docs)
	for i := 0; i < docs; i++ {
		go getSimilarity(queryVector, lsi, corpus, i, ch, &wg)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	for i := range ch {
		matches = append(matches, i)
	}

	return matches, nil
}

func getSimilarity(queryVector mat.Matrix, lsi mat.Matrix, corpus []*reddit.Comment, index int, ch chan similarComment, wg *sync.WaitGroup) {
	defer wg.Done()
	similarity := pairwise.CosineSimilarity(queryVector.(mat.ColViewer).ColView(0), lsi.(mat.ColViewer).ColView(index))
	if math.IsNaN(similarity) {
		similarity = 0
	}

	if similarity <= 0 {
		return
	}
	ch <- similarComment{
		Ranking: similarity,
		Comment: redditComment{
			Id:      corpus[index].ID,
			Comment: corpus[index].Body,
		},
	}
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	errorLogger.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
