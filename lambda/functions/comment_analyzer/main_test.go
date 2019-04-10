package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	. "io/ioutil"
)

func TestHandleRequest(t *testing.T) {
	type args struct {
		req events.APIGatewayProxyRequest
	}
	files, err := ReadDir("./tests")

	if err != nil {
		log.Fatal(err)
	}

	var tests []struct {
		name    string
		args    args
		want    events.APIGatewayProxyResponse
		wantErr bool
		comment string
	}

	for _, file := range files {
		body, _ := ReadFile("./tests/" + file.Name())
		var data []requestPayload
		err := json.Unmarshal(body, &data)
		if err != nil {
			log.Fatal(err)
		}

		for i, request := range data {
			payloadBody, _ := json.Marshal(request)

			test := struct {
				name    string
				args    args
				want    events.APIGatewayProxyResponse
				wantErr bool
				comment string
			}{
				name:    file.Name() + " " + string(i),
				comment: request.CommentToPost,
				args: args{
					req: events.APIGatewayProxyRequest{
						Resource:   "",
						Path:       "",
						HTTPMethod: "POST",
						Headers: map[string]string{
							"content-type": "application/json",
						},
						Body:            string(payloadBody),
						IsBase64Encoded: false,
					},
				},
				want: events.APIGatewayProxyResponse{
					StatusCode: http.StatusOK,
					Headers: map[string]string{
						"Access-Control-Allow-Origin": os.Getenv("CORS_ORIGIN"),
					},
				},
				wantErr: false,
			}
			tests = append(tests, test)
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleRequest(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareResponse(got, tt.want) {
				t.Errorf("Abs(-1) = %v; want %v", got, tt.want)
			}
			var body []similarComment
			err = json.Unmarshal([]byte(got.Body), &body)
			if err != nil {
				log.Fatal(err)
			}
			sort.Slice(body, func(i, j int) bool {
				return body[i].Ranking > body[j].Ranking
			})
			toLog := "Comment: " + tt.comment + "\n"
			for i, comment := range body {
				if i >= 5 {
					break
				}

				toLog += strconv.FormatFloat(comment.Ranking*100, 'f', 2, 64) + "% Similar to '" + comment.Comment.Comment + "'\n"
			}
			t.Log(toLog)
		})
	}
}

func compareResponse(r1, r2 events.APIGatewayProxyResponse) bool {
	if r1.StatusCode != r2.StatusCode {
		return false
	}
	if !reflect.DeepEqual(r1, r2) {
		return false
	}

	return true
}
