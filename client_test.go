package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testServer          = httptest.NewServer(http.HandlerFunc(SearchServer))
	aviableAccessTokens = []string{"access allowed"}
)

func SearchServer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("You accept", r.Header.Get("AccessToken"))
	for _, token := range aviableAccessTokens {
		if r.Header.Get("AccessToken") != token { // access denied
			http.Error(w, "Bad AccessToken", http.StatusUnauthorized)
			return
		}
	}
}

func TestSearchClient_FindUsers_NegativeLimit(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{
		Limit: -1,
	}

	_, err := client.FindUsers(request)
	if err == nil {
		t.Fail()
	}
}

func TestSearchClient_FindUsers_NegativeOffset(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{
		Limit: -1,
	}

	_, err := client.FindUsers(request)
	if err == nil {
		t.Fail()
	}
}

//type TestCase struct {
//	client   SearchClient
//	request  SearchRequest
//	response SearchResponse
//}

//func TestSearchClient_FindUsers(t *testing.T) {
//	cases := []TestCase{
//		{
//			client: SearchClient{
//				AccessToken: "0",
//			},
//			request: SearchRequest{
//				Limit:      0,
//				Offset:     0,
//				Query:      "",
//				OrderField: "",
//				OrderBy:    0,
//			},
//			response: SearchResponse{
//				Users:    nil,
//				NextPage: false,
//			},
//		},
//	}
//
//	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
//
//	for caseNum, item := range cases {
//		item.client.URL = testServer.URL
//		result, _ := item.client.FindUsers(item.request)
//		fmt.Printf("[%d] %v\n", caseNum, result)
//	}
//}
