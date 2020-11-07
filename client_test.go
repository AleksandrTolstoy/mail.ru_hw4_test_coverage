package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testServer         = httptest.NewServer(http.HandlerFunc(SearchServer))
	aviableAccessToken = "access allowed"
)

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != aviableAccessToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
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
	if err != nil {
		if err.Error() != "limit must be > 0" {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestSearchClient_FindUsers_NegativeOffset(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{
		Offset: -1,
	}

	_, err := client.FindUsers(request)
	if err != nil {
		if err.Error() != "offset must be > 0" {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestSearchClient_FindUsers_AccessDenied(t *testing.T) {
	client := SearchClient{
		AccessToken: "access denied",
		URL:         testServer.URL,
	}

	request := SearchRequest{}

	_, err := client.FindUsers(request)
	if err != nil {
		if err.Error() != "Bad AccessToken" {
			t.Fail()
		}
	} else {
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
