package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

var (
	testServer = httptest.NewServer(http.HandlerFunc(SearchServer))
)

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != "access allowed" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	xmlData, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	users := decodeUsers(xmlData)
	fmt.Println(users)
}

func decodeUsers(xmlData []byte) []User {
	input := bytes.NewReader(xmlData)
	decoder := xml.NewDecoder(input)

	curUserIdx := -1
	users := make([]User, 0)
	var structField string
	for {
		tok, tokenErr := decoder.Token()
		if tokenErr == io.EOF {
			break
		}

		if tok, ok := tok.(xml.StartElement); ok {
			switch {
			case tok.Name.Local == "row":
				users = append(users, User{})
				curUserIdx++
			case tok.Name.Local == "first_name":
				if err := decoder.DecodeElement(&structField, &tok); err != nil {
					fmt.Printf("first_name: error : %v\n", err)
				}

				u := &users[curUserIdx]
				u.Name = structField
			case tok.Name.Local == "last_name":
				if err := decoder.DecodeElement(&structField, &tok); err != nil {
					fmt.Printf("last_name: error : %v\n", err)
				}

				u := &users[curUserIdx]
				u.Name += " " + structField
			default:
				for _, otherField := range []string{"id", "age", "about", "gender"} {
					if tok.Name.Local == otherField {
						if err := decoder.DecodeElement(&structField, &tok); err != nil {
							fmt.Printf("%s: error : %v\n", otherField, err)
						}

						u := &users[curUserIdx]
						switch {
						case otherField == "id":
							id, _ := strconv.Atoi(structField)
							u.Id = id
						case otherField == "age":
							age, _ := strconv.Atoi(structField)
							u.Age = age
						case otherField == "about":
							u.About = structField
						case otherField == "gender":
							u.Gender = structField
						}
					}
				}
			}
		}
	}

	return users
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

func TestSearchClient_FindUsers_Normal(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{}

	_, _ = client.FindUsers(request)
}
