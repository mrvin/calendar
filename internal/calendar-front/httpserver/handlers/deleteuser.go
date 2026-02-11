package handlers

import (
	"log/slog"
	"net/http"
)

func (h *Handler) DeleteUser(res http.ResponseWriter, req *http.Request) {
	accessToken := req.URL.Query().Get("jwt-token")
	if accessToken == "" {
		slog.Error("Cant get jwt token")
		h.ErrMsg(res)
		return
	}
	if err := h.client.DeleteUser(req.Context(), accessToken); err != nil {
		slog.Error("Delete user: " + err.Error())
		return
	}

	text := resp{Title: "Delete user", Body: struct{ Text string }{"User deleted successfully"}}
	if err := h.templates.Execute("text.html", res, text); err != nil {
		slog.Error("Execute delete user template: " + err.Error())
		h.ErrMsg(res)
		return
	}
}
