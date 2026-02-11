package handlers

import (
	"log/slog"
	"net/http"
)

func (h *Handler) DeleteEvent(res http.ResponseWriter, req *http.Request) {
	idEvent, err := getID(req)
	if err != nil {
		slog.Error("Get id event: " + err.Error())
		h.ErrMsg(res)
		return
	}
	accessToken := req.URL.Query().Get("jwt-token")
	if err := h.client.DeleteEvent(req.Context(), accessToken, idEvent); err != nil {
		slog.Error("Delete event: " + err.Error())
		return
	}

	text := resp{Title: "Delete event", Body: struct{ Text string }{"Event deleted successfully"}}
	if err := h.templates.Execute("text.html", res, text); err != nil {
		slog.Error("Execute delete event template: " + err.Error())
		h.ErrMsg(res)
		return
	}
}
