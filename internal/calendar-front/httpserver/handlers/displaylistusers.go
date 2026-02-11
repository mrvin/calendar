package handlers

import (
	"log/slog"
	"net/http"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
)

func (h *Handler) DisplayListUsers(res http.ResponseWriter, req *http.Request) {
	users, err := h.client.ListUsers(req.Context(), "")
	if err != nil {
		slog.Error("List users: " + err.Error())
		return
	}

	dataUsers := struct {
		Title string
		Body  struct {
			Users []storage.User
		}
	}{
		Title: "List users",
		Body: struct {
			Users []storage.User
		}{users},
	}

	if err := h.templates.Execute("list-users.html", res, dataUsers); err != nil {
		slog.Error("Execute display list users template: " + err.Error())
		h.ErrMsg(res)
		return
	}
}
