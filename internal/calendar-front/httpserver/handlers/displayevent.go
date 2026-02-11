package handlers

import (
	"log/slog"
	"net/http"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
)

func (h *Handler) DisplayEvent(res http.ResponseWriter, req *http.Request) {
	idEvent, err := getID(req)
	if err != nil {
		slog.Error("Get id event: " + err.Error())
		h.ErrMsg(res)
		return
	}
	accessToken := req.URL.Query().Get("jwt-token")
	event, err := h.client.GetEvent(req.Context(), accessToken, idEvent)
	if err != nil {
		slog.Error("Get event: " + err.Error())
		return
	}

	dataEvent := struct {
		Title string
		Body  struct {
			AccessToken string
			Event       storage.Event
		}
	}{
		Title: "Event",
		Body: struct {
			AccessToken string
			Event       storage.Event
		}{accessToken, *event},
	}

	if err := h.templates.Execute("event.html", res, dataEvent); err != nil {
		slog.Error("Execute event template: " + err.Error())
		h.ErrMsg(res)
		return
	}
}
