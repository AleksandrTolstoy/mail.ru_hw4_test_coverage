package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var testServer = httptest.NewServer(http.HandlerFunc(SearchServer))

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
	decodedUsers := decodeUsers(xmlData)

	q := r.URL.Query()

	offset, _ := strconv.Atoi(q.Get("offset"))
	// Special case. Actually i wouldn't do that in real project
	// This error needs to be handled in the home work (1)
	if offset >= 15 && offset <= 24 {
		q.Set("offset", strconv.Itoa(offset+1))
		r.URL.RawQuery = q.Encode()
		http.Redirect(w, r, r.URL.RequestURI(), http.StatusTemporaryRedirect)
		return
	}

	// Special case. Actually i wouldn't do that in real project
	// This error needs to be handled in the home work (2)
	if offset == 25 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Special case. Actually i wouldn't do that in real project
	// This error needs to be handled in the home work (3)
	if offset == 26 {
		<-r.Context().Done()
		return
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	orderBy, _ := strconv.Atoi(q.Get("order_by"))
	query := q.Get("query")
	orderField := q.Get("order_field")

	if !isOrderAvailable(orderBy) {
		w.WriteHeader(http.StatusBadRequest)
		// Special case. Actually i wouldn't do that in real project
		// This error needs to be handled in the home work (4)
		w.Write([]byte("Unknown order"))
		return
	}

	searchResult := searchUsers(query, limit, decodedUsers)
	if len(searchResult) == 0 {
		// Special case. Actually i wouldn't do that in real project
		// This error needs to be handled in the home work (5)
		w.Write([]byte("No users found!"))
		return
	}

	statusCode := sortUsers(orderField, orderBy, searchResult)

	if statusCode == http.StatusBadRequest {
		w.WriteHeader(statusCode)
		errResp, _ := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})
		w.Write(errResp)
		return
	}

	if offset >= len(searchResult) {
		w.WriteHeader(http.StatusBadRequest)
		errResp, _ := json.Marshal(SearchErrorResponse{Error: "no items with this offset"})
		w.Write(errResp)
		return
	}

	searchResult = searchResult[offset:]
	resp, err := json.Marshal(searchResult)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(resp)
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

func isOrderAvailable(orderBy int) bool {
	for _, order := range []int{OrderByAsc, OrderByAsIs, OrderByDesc} {
		if orderBy == order {
			return true
		}
	}

	return false
}

func searchUsers(query string, limit int, decodedUsers []User) []User {
	if query != "" {
		result := make([]User, 0, limit)
		for _, u := range decodedUsers {
			if len(result) == limit {
				break
			}

			if strings.Contains(u.Name, query) || strings.Contains(u.About, query) {
				result = append(result, u)
			}
		}

		return result
	}

	return decodedUsers
}

func sortUsers(orderField string, orderBy int, users []User) (statusCode int) {
	switch orderField {
	case "", "Name":
		if orderBy == OrderByAsIs {
			break
		}

		sort.Slice(users, func(i, j int) bool {
			if orderBy == OrderByAsc {
				return users[i].Name > users[j].Name
			}

			return users[i].Name < users[j].Name
		})
	case "Id":
		if orderBy == OrderByAsIs {
			break
		}

		sort.Slice(users, func(i, j int) bool {
			if orderBy == OrderByAsc {
				return users[i].Id > users[j].Id
			}

			return users[i].Id < users[j].Id
		})
	case "Age":
		if orderBy == OrderByAsIs {
			break
		}

		sort.Slice(users, func(i, j int) bool {
			if orderBy == OrderByAsc {
				return users[i].Age > users[j].Age
			}

			return users[i].Age < users[j].Age
		})
	default:
		return http.StatusBadRequest
	}

	return http.StatusOK
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
	testSearchClient := SearchClient{
		AccessToken: "access denied", // note !
		URL:         testServer.URL,
	}

	_, err := testSearchClient.FindUsers(SearchRequest{})
	if err != nil {
		if err.Error() != "Bad AccessToken" {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestSearchClient_FindUsers_TooManyRedirects(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{
		Offset: 15,
	}

	_, err := client.FindUsers(request)
	if err != nil {
		if !strings.Contains(err.Error(), "stopped after 10 redirects") {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestSearchClient_FindUsers_BigOffset(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{
		Offset:     100500,
		OrderField: "Id",
	}

	_, err := client.FindUsers(request)
	if err != nil {
		if !strings.Contains(err.Error(), "unknown bad request error: ") {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestSearchClient_FindUsers_BadOrderField(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{
		OrderField: "Random",
	}

	_, err := client.FindUsers(request)
	if err != nil {
		if err.Error() != fmt.Sprintf("OrderFeld %s invalid", request.OrderField) {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestSearchClient_FindUsers_UnknownOrder(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{
		OrderBy: 100500,
	}

	_, err := client.FindUsers(request)
	if err != nil {
		if !strings.Contains(err.Error(), "cant unpack error json: ") {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestSearchClient_FindUsers_BigLimit(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{
		Limit:      100,
		Offset:     0,
		Query:      "a",
		OrderField: "Id",
		OrderBy:    OrderByAsc,
	}

	_, err := client.FindUsers(request)
	if err != nil {
		t.Fail()
	}
}

func TestSearchClient_FindUsers_NoUsersFound(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{
		Limit:      5,
		Offset:     2,
		Query:      "I'm the best man in the world", // haha
		OrderField: "Id",
		OrderBy:    OrderByAsc,
	}

	_, err := client.FindUsers(request)
	if err != nil {
		if !strings.Contains(err.Error(), "cant unpack result json: ") {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestSearchClient_FindUsers_ISE(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{
		Offset: 25,
	}

	_, err := client.FindUsers(request)
	if err != nil {
		if err.Error() != "SearchServer fatal error" {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestSearchClient_FindUsers_Timeout(t *testing.T) {
	client := SearchClient{
		AccessToken: "access allowed",
		URL:         testServer.URL,
	}

	request := SearchRequest{
		Offset: 26,
	}

	_, err := client.FindUsers(request)
	if err != nil {
		if !strings.Contains(err.Error(), "timeout for ") {
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

	request := SearchRequest{
		Limit:      4,
		Offset:     1,
		Query:      "a",
		OrderField: "Id",
		OrderBy:    OrderByAsc,
	}

	_, err := client.FindUsers(request)
	if err != nil {
		t.Fail()
	}
}
