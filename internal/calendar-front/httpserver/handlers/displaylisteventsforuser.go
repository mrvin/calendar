package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
)

func (h *Handler) DisplayListEventsForUser(res http.ResponseWriter, req *http.Request) {
	var err error
	days := 0
	daysStr := req.URL.Query().Get("days")
	if daysStr != "" {
		days, err = strconv.Atoi(daysStr)
		if err != nil {
			slog.Error("Convert days: " + err.Error())
			return
		}
	}
	accessToken := req.PathValue("accessToken")
	fmt.Printf("Access token: %s (ListEvents)\n", accessToken)
	events, err := h.client.ListEventsForUser(req.Context(), accessToken, days)
	if err != nil {
		slog.Error("ListEventsForUser: " + err.Error())
		h.ErrMsg(res)
		return
	}
	fmt.Printf("Events: %v\n", events)
	dataEvent := struct {
		Title string
		Body  struct {
			AccessToken string
			Events      []storage.Event
		}
	}{
		Title: "List events",
		Body: struct {
			AccessToken string
			Events      []storage.Event
		}{accessToken, events},
	}

	if err := h.templates.Execute("list-events.html", res, dataEvent); err != nil {
		slog.Error("Execute display list events template: " + err.Error())
		h.ErrMsg(res)
		return
	}
}
