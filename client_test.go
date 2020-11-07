package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func SearchServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "You accept", r.Header.Get("AccessToken"))
}

type TestCase struct {
	client   SearchClient
	request  SearchRequest
	response SearchResponse
}

func TestSearchClient_FindUsers(t *testing.T) {
	cases := []TestCase{
		{
			client: SearchClient{
				AccessToken: "0",
			},
			request: SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    0,
			},
			response: SearchResponse{
				Users:    nil,
				NextPage: false,
			},
		},
	}

	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, item := range cases {
		item.client.URL = testServer.URL
		result, _ := item.client.FindUsers(item.request)
		fmt.Printf("[%d] %v\n", caseNum, result)
	}
}
