package handlers

import (
	"log/slog"
	"net/http"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
)

func (h *Handler) DisplayUser(res http.ResponseWriter, req *http.Request) {
	user, err := h.client.GetUser(req.Context(), "")
	if err != nil {
		slog.Error("Get user: " + err.Error())
		return
	}

	dataUser := struct {
		Title string
		Body  struct {
			User *storage.User
		}
	}{
		Title: "User",
		Body: struct {
			User *storage.User
		}{user},
	}

	if err := h.templates.Execute("user.html", res, dataUser); err != nil {
		slog.Error("Execute display user template: " + err.Error())
		h.ErrMsg(res)
		return
	}
}
