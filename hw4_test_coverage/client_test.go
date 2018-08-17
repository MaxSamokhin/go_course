package main

import (
    "testing"
    "net/http/httptest"
    "net/http"
    "strconv"
    "io/ioutil"
    "encoding/xml"
    "os"
    "encoding/json"
    "strings"
    "sort"
    "fmt"
)

const DATASET = "dataset.xml"
const ACCESS_TOKEN_GOOD = "AccessTokenGood"
const OrderByError = "OrderByError"
const LimitError = "LimitError"
const OffsetError = "OffsetError"

type UserDataset struct {
    Id        int    `xml:"id"`
    FirstName string `xml:"first_name"`
    LastName  string `xml:"last_name"`
    Age       int    `xml:"age"`
    About     string `xml:"about"`
    Gender    string `xml:"gender"`
}

type UserData struct {
    Users []UserDataset `xml:"row"`
}

type ByID []User
type ByAge []User
type ByName []User

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return a[i].Id < a[j].Id }

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return strings.Compare(a[i].Name, a[j].Name) < 0 }

func getUrlParam(r *http.Request) (*SearchRequest, string) {
    req := new(SearchRequest)
    var err error = nil

    req.Limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
    if err != nil {
        return nil, LimitError
    }

    req.Offset, err = strconv.Atoi(r.FormValue("offset"))
    if err != nil {
        return nil, OffsetError
    }

    req.Query = r.FormValue("query")
    req.OrderField = r.FormValue("order_field")
    if req.OrderField != "" && strings.Index("Id_Name_Age", req.OrderField) < 0 {
        return nil, ErrorBadOrderField
    }

    req.OrderBy, err = strconv.Atoi(r.FormValue("order_by"))
    if err != nil {
        return nil, OrderByError
    }

    return req, ""
}

func getUsersFromDataset() *[]User {
    file, err := os.Open(DATASET)
    if err != nil {
        panic("error opening file")
    }
    defer file.Close()

    xmlFile, err := ioutil.ReadAll(file)
    if err != nil {
        panic("error open file")
    }

    dataset := new(UserData)
    err = xml.Unmarshal(xmlFile, &dataset)
    if err != nil {
        panic("error Unmarshal xml")
    }

    users := make([]User, len(dataset.Users))

    for index, elem := range dataset.Users {
        users[index].Id = elem.Id
        users[index].Name = elem.FirstName + " " + elem.LastName
        users[index].Age = elem.Age
        users[index].About = elem.About
        users[index].Gender = elem.Gender
    }

    return &users
}

func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
    dataJSON, err := json.Marshal(data)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    w.WriteHeader(statusCode)
    w.Write(dataJSON)
}

func filterByFields(query string, users []User) *[]User {
    result := make([]User, 0)

    for _, v := range users {
        if strings.Contains(v.Name, query) || strings.Contains(v.About, query) {
            result = append(result, v)
        }
    }

    return &result
}

func sortByField(orderField string, orderBy int, users []User) *[]User {

    switch orderField {
    case "id":
        if orderBy == OrderByDesc {
            sort.Sort(ByID(users))
            break
        }
        sort.Sort(sort.Reverse(ByID(users)))
    case "age":
        if orderBy == OrderByDesc {
            sort.Sort(ByAge(users))
            break
        }
        sort.Sort(sort.Reverse(ByAge(users)))
    case "name":
        if orderBy == OrderByDesc {
            sort.Sort(ByName(users))
            break
        }
        sort.Sort(sort.Reverse(ByAge(users)))
    case "":
        if orderBy == OrderByDesc {
            sort.Sort(ByName(users))
            break
        }
        sort.Sort(sort.Reverse(ByAge(users)))
    }

    return &users
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
    request, err := getUrlParam(r)
    if err != "" {
        sendJSON(w, SearchErrorResponse{Error: "ErrorBadOrderField"},http.StatusBadRequest)
        return
    }

    token := r.Header.Get("AccessToken")
    if token != ACCESS_TOKEN_GOOD {
        sendJSON(w, nil, http.StatusUnauthorized)
        return
    }

    users := *getUsersFromDataset()

    if request.Query != "" {
        users = *filterByFields(request.Query, users)
    }

    if request.OrderBy != OrderByAsIs {
        users = *sortByField(request.OrderField, request.OrderBy, users)
    }

    endIndex := len(users)
    if len(users) > request.Limit+request.Offset {
        endIndex = request.Limit + request.Offset
    }

    users = users[request.Offset:endIndex]
    sendJSON(w, users, http.StatusOK)
}

func TestGoodGetResultMaxLimit(t *testing.T) {
    testCase := SearchRequest{
        Limit:      45,
        Offset:     0,
        Query:      "Mayer",
        OrderField: "Id",
        OrderBy:    OrderByAsc,
    }

    ts := httptest.NewServer(http.HandlerFunc(SearchServer))
    defer ts.Close()

    client := &SearchClient{
        AccessToken: "AccessTokenGood",
        URL:         ts.URL,
    }

    result, err := client.FindUsers(testCase)
    if err != nil {
        t.Error("Unexpected error")
    }

    if len(result.Users) != 1 {
        t.Errorf("Wrong number of users: got %+v, want %+v", len(result.Users), 1)
    }
}

func TestBadIncorrectLimitAndOffset(t *testing.T) {
    cases := []SearchRequest{
        SearchRequest{
            Limit:      -10,
            Offset:     0,
            Query:      "Mayer",
            OrderField: "Id",
            OrderBy:    OrderByAsc,
        },
        SearchRequest{
            Limit:      1,
            Offset:     -10,
            Query:      "Mayer",
            OrderField: "Id",
            OrderBy:    OrderByAsc,
        },
    }

    ts := httptest.NewServer(http.HandlerFunc(SearchServer))
    defer ts.Close()

    for _, item := range cases {
        client := &SearchClient{
            AccessToken: "AccessTokenGood",
            URL:         ts.URL,
        }

        result, err := client.FindUsers(item)

        if result != nil || err == nil {
            t.Errorf("Incorrect Limit or Offset")
        }
    }
}

func TestTimeout(t *testing.T) {
    testCase := SearchRequest{
        Limit:      45,
        Offset:     0,
        Query:      "Mayer",
        OrderField: "Id",
        OrderBy:    OrderByAsc,
    }

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        for i := 0; i < 10; i-- {
        }
    }))

    client := SearchClient{
        AccessToken: "AccessTokenGood",
        URL:         ts.URL,
    }

    result, err := client.FindUsers(testCase)

    if result != nil || err == nil {
        t.Errorf("expected error")
    }
}

func TestReturnInternalServerError(t *testing.T) {
    testCase := SearchRequest{
        Limit:      5,
        Offset:     0,
        Query:      "Mayer",
        OrderField: "Id",
        OrderBy:    OrderByAsc,
    }

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError) // 500
    }))

    client := SearchClient{
        AccessToken: "AccessTokenGood",
        URL:         ts.URL,
    }

    result, err := client.FindUsers(testCase)

    if result != nil || err == nil {
        t.Errorf("expected error")
    }
}

func TestReturnUnauthorized(t *testing.T) {
    testCase := SearchRequest{
        Limit:      5,
        Offset:     0,
        Query:      "Mayer",
        OrderField: "Id",
        OrderBy:    OrderByAsc,
    }

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusUnauthorized)
    }))

    client := SearchClient{
        AccessToken: "AccessTokenGood",
        URL:         ts.URL,
    }

    result, err := client.FindUsers(testCase)

    if result != nil || err == nil {
        t.Errorf("expected error")
    }
}

func TestReturnBadRequest(t *testing.T) {
    testCase := SearchRequest{
        Limit:      5,
        Offset:     0,
        Query:      "Mayer",
        OrderField: "Id",
        OrderBy:    OrderByAsc,
    }

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
    }))

    client := SearchClient{
        AccessToken: "AccessTokenGood",
        URL:         ts.URL,
    }

    result, err := client.FindUsers(testCase)

    if result != nil || err == nil {
        t.Errorf("expected error")
    }
}

func TestUncorrectJSON(t *testing.T) {
    testCase := SearchRequest{
        Limit:      5,
        Offset:     0,
        Query:      "Mayer",
        OrderField: "Id",
        OrderBy:    OrderByAsc,
    }

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("{\"Id\":''2'2}"))
    }))

    client := SearchClient{
        AccessToken: "AccessTokenGood",
        URL:         ts.URL,
    }

    result, err := client.FindUsers(testCase)

    if result != nil || err == nil {
        t.Errorf("expected error")
    }
}

func TestNextPage(t *testing.T) {
    testCase := SearchRequest{
        Limit:      1,
        Offset:     0,
        Query:      "",
        OrderField: "Id",
        OrderBy:    OrderByAsc,
    }

    ts := httptest.NewServer(http.HandlerFunc(SearchServer))
    defer ts.Close()

    client := &SearchClient{
        AccessToken: "AccessTokenGood",
        URL:         ts.URL,
    }

    result, err := client.FindUsers(testCase)
    if err != nil {
        t.Error("Unexpected error")
    }

    if len(result.Users) != 1 {
        t.Errorf("Wrong number of users: got %+v, want %+v", len(result.Users), 1)
    }
}

func TestReturnError(t *testing.T) {
    testCase := SearchRequest{
        Limit:      5,
        Offset:     0,
        Query:      "Mayer",
        OrderField: "Id",
        OrderBy:    OrderByAsc,
    }

    client := SearchClient{
        AccessToken: "AccessTokenGood",
        URL:         "noUrl",
    }

    result, err := client.FindUsers(testCase)

    if result != nil || err == nil {
        t.Errorf("expected error")
    }
}

func TestBadOrderField(t *testing.T) {
    testCase := SearchRequest{
        Limit:      5,
        Offset:     0,
        Query:      "Mayer",
        OrderField: "bad",
        OrderBy:    OrderByAsc,
    }

    ts := httptest.NewServer(http.HandlerFunc(SearchServer))

    client := SearchClient{
        AccessToken: "AccessTokenGood",
        URL:         ts.URL,
    }

    result, err := client.FindUsers(testCase)
    fmt.Println(err, result)
    if result != nil || err == nil {
        t.Errorf("expected error")
    }
}

func TestUnknownBadRequest(t *testing.T) {
    testCase := SearchRequest{
        Limit:      5,
        Offset:     0,
        Query:      "Mayer",
        OrderField: "bad",
        OrderBy:    OrderByAsc,
    }

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sendJSON(w, SearchErrorResponse{Error: "NewError"}, http.StatusBadRequest)
    }))

    client := SearchClient{
        AccessToken: "AccessTokenGood",
        URL:         ts.URL,
    }

    result, err := client.FindUsers(testCase)
    fmt.Println(err, result)
    if result != nil || err == nil {
        t.Errorf("expected error")
    }
}
