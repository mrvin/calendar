/*
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpserver "github.com/mrvin/calendar/internal/calendar/server/http"
	handlerevent "github.com/mrvin/calendar/internal/calendar/server/http/handlers/event"
	handlerusersignin "github.com/mrvin/calendar/internal/calendar/server/http/handlers/user/signin"
	handlerusersignup "github.com/mrvin/calendar/internal/calendar/server/http/handlers/user/signup"
	handleruserupdate "github.com/mrvin/calendar/internal/calendar/server/http/handlers/user/update"
	authservice "github.com/mrvin/calendar/internal/calendar/service/auth"
	eventservice "github.com/mrvin/calendar/internal/calendar/service/event"
	"github.com/mrvin/calendar/internal/storage"
	memorystorage "github.com/mrvin/calendar/internal/storage/memory"
	sqlstorage "github.com/mrvin/calendar/internal/storage/sql"
)

//nolint:tagliatelle
type User struct {
	Name     string
	Password string
	Email    string
	Token    string
}

const urlSignup = "http://localhost:8080/signup"
const urlLogin = "http://localhost:8080/login"
const urlUsers = "http://localhost:8080/user"
const urlEvents = "http://localhost:8080/event"

const contextTimeoutDB = 2 * time.Second

var confDBTest = sqlstorage.Conf{"postgres", "postgres", 5432, "event-db", "event-db", "event-db"}

func initServerHTTP(st storage.Storage) *httpserver.Server {
	conf := httpserver.Conf{"localhost", 8080, false, httpserver.ConfHTTPS{}}
	confAuth := authservice.Conf{"secret key", 15}
	authService := authservice.New(st, &confAuth)
	eventService := eventservice.New(st)
	server := httpserver.New(&conf, authService, eventService)

	return server
}

func TestHandleUserMemory(t *testing.T) {
	st := memorystorage.New()
	server := initServerHTTP(st)
	testHandleUser(t, server)
}

func TestHandleEventMemory(t *testing.T) {
	st := memorystorage.New()
	server := initServerHTTP(st)
	testHandleEvent(t, server)
}

func TestHandleUserSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDB)
	defer cancel()
	st, err := sqlstorage.New(ctx, &confDBTest)
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer st.Close()

	server := initServerHTTP(st)
	testHandleUser(t, server)
}

func TestHandleEventSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDB)
	defer cancel()
	st, err := sqlstorage.New(ctx, &confDBTest)
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer st.Close()

	server := initServerHTTP(st)
	testHandleEvent(t, server)
}

func testHandleUser(t *testing.T, server *httpserver.Server) {
	users := []User{
		{
			Name:     "Howard Mendoza",
			Password: "qwerty",
			Email:    "Howard.Mendoza@mail.com",
		},
		{
			Name:     "Brian Olson",
			Password: "Olson123456",
			Email:    "B.Olson@gmail.com",
		},
		{
			Name:     "Clarence Olson",
			Password: "Russian Bear",
			Email:    "Clarence.Olson@yandex.ru",
		},
	}

	// Create users
	for i, user := range users {
		testHandleSignUp(t, server, &user, http.StatusCreated)
		users[i].Token = testHandleLogin(t, server, &user, http.StatusOK)
	}

	// Get users
	for i, user := range users {
		user := testHandleGetUser(t, server, &user, http.StatusOK)
		if user != nil {
			if user.Name != users[i].Name {
				t.Errorf("mismatch name user: %s, %s", user.Name, users[i].Name)
			}
			if user.Email != users[i].Email {
				t.Errorf("mismatch email user: %s, %s", user.Email, users[i].Email)
			}
		}
	}

	// Update user email and get user
	for i, user := range users {
		users[i].Email = strings.ToLower(user.Email)
		testHandleUpdateUser(t, server, &users[i], http.StatusOK)

		user := testHandleGetUser(t, server, &user, http.StatusOK)
		if user != nil {
			if user.Email != users[i].Email {
				t.Errorf("mismatch email user: %s, %s", user.Email, users[i].Email)
			}
		}
	}

	// Delete all users and trying get, update, delete user that doesn't exist
	for _, user := range users {
		testHandleDeleteUser(t, server, &user, http.StatusOK)

		testHandleGetUser(t, server, &user, http.StatusBadRequest)
		testHandleUpdateUser(t, server, &user, http.StatusBadRequest)
		testHandleDeleteUser(t, server, &user, http.StatusBadRequest)
	}
}

func testHandleEvent(t *testing.T, server *httpserver.Server) {
	user := User{
		Name:     "Bob",
		Password: "qwerty",
		Email:    "bobi@mail.com",
	}

	user.ID = testHandleSignUp(t, server, &user, http.StatusCreated)
	user.Token = testHandleLogin(t, server, &user, http.StatusOK)

	events := []storage.Event{
		{Title: "Bob's Birthday", Description: "Birthday February 24, 1993. Party in nature.",
			StartTime: time.Date(2022, time.February, 27, 10, 0, 0, 0, time.UTC),
			StopTime:  time.Date(2022, time.February, 27, 23, 0, 0, 0, time.UTC), UserID: user.ID},
		{Title: "Alis's Birthday", Description: "Birthday April 12, 1996. House party",
			StartTime: time.Date(2022, time.April, 13, 19, 0, 0, 0, time.UTC),
			StopTime:  time.Date(2022, time.April, 13, 21, 0, 0, 0, time.UTC), UserID: user.ID},
		{Title: "Jim's Birthday", Description: "Birthday August 15, 1994. Party at the restaurant",
			StartTime: time.Date(2022, time.August, 17, 16, 0, 0, 0, time.UTC),
			StopTime:  time.Date(2022, time.August, 17, 19, 0, 0, 0, time.UTC), UserID: user.ID},
		{Title: "Bill's Birthday", Description: "Birthday November 6, 1990. Party at the club",
			StartTime: time.Date(2022, time.November, 7, 11, 0, 0, 0, time.UTC),
			StopTime:  time.Date(2022, time.November, 7, 12, 0, 0, 0, time.UTC), UserID: user.ID},
	}

	// Create events
	for i := range events {
		events[i].ID = testHandleCreateEvent(t, server, user.Token, &events[i], http.StatusCreated)
	}

	// Get events
	for i := range events {
		event := testHandleGetEvent(t, server, user.Token, events[i].ID, http.StatusOK)
		if event != nil {
			if event.Title != events[i].Title {
				t.Errorf("mismatch title event: %s, %s", event.Title, events[i].Title)
			}
			if event.Description != events[i].Description {
				t.Errorf("mismatch description event: %s, %s", event.Description, events[i].Description)
			}
		}
	}

	// Update event title and get event
	for i := range events {
		events[i].Title = strings.ToUpper(events[i].Title)
		testHandleUpdateEvent(t, server, user.Token, &events[i], http.StatusOK)

		event := testHandleGetEvent(t, server, user.Token, events[i].ID, http.StatusOK)
		if event != nil {
			if event.Title != events[i].Title {
				t.Errorf("mismatch title event: %s, %s", event.Title, events[i].Title)
			}
		}
	}

	// Delete all event without last and trying get, update, delete event that doesn't exist
	for i := 0; i < len(events)-1; i++ {
		testHandleDeleteEvent(t, server, user.Token, events[i].ID, http.StatusOK)

		testHandleGetEvent(t, server, user.Token, events[i].ID, http.StatusBadRequest)
		testHandleUpdateEvent(t, server, user.Token, &events[i], http.StatusBadRequest)
		testHandleDeleteEvent(t, server, user.Token, events[i].ID, http.StatusBadRequest)
	}

	testHandleDeleteEvent(t, server, user.Token, int64(len(events)), http.StatusOK)
	testHandleDeleteUser(t, server, &user, http.StatusOK)
	testHandleDeleteEvent(t, server, user.Token, int64(len(events)), http.StatusBadRequest)
}

func testHandleSignUp(t *testing.T, server *httpserver.Server, user *User, status int) {
	res := httptest.NewRecorder()

	request := handlerusersignup.RequestSignUp{
		UserName: user.Name,
		Password: user.Password,
		Email:    user.Email,
	}
	dataJsonRequest, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("HandleCreateUser: cant marshal JSON: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, urlSignup, bytes.NewReader(dataJsonRequest))
	if err != nil {
		t.Fatalf("HandleCreateUser: create request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")

	server.Handler.ServeHTTP(res, req)

	if res.Code != status {
		t.Errorf("HandleCreateUser: response code is %d; want: %d", res.Code, status)
	}

	var response handlerusersignup.ResponseSignUp
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("HandleCreateUser: cant unmarshal JSON: %v", err)
	}

	return response.ID
}

func testHandleLogin(t *testing.T, server *httpserver.Server, user *User, status int) string {
	res := httptest.NewRecorder()

	request := handlerusersignin.RequestSignIn{
		UserName: user.Name,
		Password: user.Password,
	}

	dataJsonRequest, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("HandleLogin: cant marshal JSON: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, urlLogin, bytes.NewReader(dataJsonRequest))
	if err != nil {
		t.Fatalf("HandleLogin: create request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")

	server.Handler.ServeHTTP(res, req)

	if res.Code != status {
		t.Errorf("HandleLogin: response code is %d; want: %d", res.Code, status)
	}

	var response handlerusersignin.ResponseSignIn
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("HandleLogin: cant unmarshal JSON: %v", err)
	}

	return response.AccessToken
}

func testHandleGetUser(t *testing.T, server *httpserver.Server, user *User, status int) *storage.User {
	res := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, urlUsers, nil)
	if err != nil {
		t.Fatalf("HandleGetUser: create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+user.Token)

	server.Handler.ServeHTTP(res, req)

	if res.Code != status {
		t.Errorf("HandleGetUser: response code is %d; want: %d", res.Code, status)
	}

	var stUser storage.User
	if err := json.Unmarshal(res.Body.Bytes(), &stUser); err != nil {
		t.Fatalf("HandleGetUser: cant unmarshal JSON: %v", err)
	}

	return &stUser
}

func testHandleUpdateUser(t *testing.T, server *httpserver.Server, user *User, status int) {
	res := httptest.NewRecorder()

	request := handleruserupdate.RequestUpdateUser{
		UserName: user.Name,
		Password: user.Password,
		Email:    user.Email,
	}
	dataJsonRequest, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("HandleUpdateUser: cant marshal JSON: %v", err)
	}
	req, err := http.NewRequest(http.MethodPut, urlUsers, bytes.NewReader(dataJsonRequest))
	if err != nil {
		t.Fatalf("HandleUpdateUser: create request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user.Token)

	server.Handler.ServeHTTP(res, req)

	if res.Code != status {
		t.Errorf("HandleUpdateUser: response code is %d; want: %d", res.Code, status)
	}
}

func testHandleDeleteUser(t *testing.T, server *httpserver.Server, user *User, status int) {
	res := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodDelete, urlUsers, nil)
	if err != nil {
		t.Fatalf("HandleDeleteUser: create request %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+user.Token)

	server.Handler.ServeHTTP(res, req)

	if res.Code != status {
		t.Errorf("HandleDeleteUser: response code is %d; want: %d", res.Code, status)
	}
}

func testHandleCreateEvent(t *testing.T, server *httpserver.Server, token string, event *storage.Event, status int) int64 {
	res := httptest.NewRecorder()

	dataJsonEvent, err := json.Marshal(*event)
	if err != nil {
		t.Fatalf("HandleCreateEvent: cant marshal JSON: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, urlEvents, bytes.NewReader(dataJsonEvent))
	if err != nil {
		t.Fatalf("HandleCreateEvent: create request %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	server.Handler.ServeHTTP(res, req)

	if res.Code != status {
		t.Errorf("HandleCreateEvent: response code is %d; want: %d", res.Code, status)
	}

	var response handlerevent.ResponseCreateEvent
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("HandleCreateEvent: cant unmarshal JSON: %v", err)
	}

	return response.ID
}

func testHandleGetEvent(t *testing.T, server *httpserver.Server, token string, id int64, status int) *storage.Event {
	res := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?id=%d", urlEvents, id), nil)
	if err != nil {
		t.Fatalf("HandleGetEvent: create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	server.Handler.ServeHTTP(res, req)

	if res.Code != status {
		t.Errorf("HandleGetEvent: response code is %d; want: %d", res.Code, status)
	}

	var event storage.Event
	if err := json.Unmarshal(res.Body.Bytes(), &event); err != nil {
		t.Errorf("HandleGetEvent: cant unmarshal JSON: %v", err)
	}

	return &event
}

func testHandleUpdateEvent(t *testing.T, server *httpserver.Server, token string, event *storage.Event, status int) {
	res := httptest.NewRecorder()

	dataJsonEvent, err := json.Marshal(*event)
	if err != nil {
		t.Fatalf("HandleUpdateEvent: cant marshal JSON: %v", err)
	}
	req, err := http.NewRequest(http.MethodPut, urlEvents, bytes.NewReader(dataJsonEvent))
	if err != nil {
		t.Fatalf("HandleUpdateEvent: create request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	server.Handler.ServeHTTP(res, req)

	if res.Code != status {
		t.Errorf("HandleUpdateEvent: response code is %d; want: %d", res.Code, status)
	}
}

func testHandleDeleteEvent(t *testing.T, server *httpserver.Server, token string, id int64, status int) {
	res := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s?id=%d", urlEvents, id), nil)
	if err != nil {
		t.Fatalf("HandleDeleteEvent: create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	server.Handler.ServeHTTP(res, req)

	if res.Code != status {
		t.Errorf("HandleDeleteEvent: response code is %d; want: %d", res.Code, status)
	}
}
*/