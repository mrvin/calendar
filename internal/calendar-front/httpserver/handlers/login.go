package handlers

import (
	"log/slog"
	"net/http"
)

func (h *Handler) DisplayFormLogin(res http.ResponseWriter, req *http.Request) {
	data := resp{Title: "Login user"}
	if err := h.templates.Execute("form-login.html", res, data); err != nil {
		slog.Error("Execute display form user login template: " + err.Error())
		h.ErrMsg(res)
		return
	}
}

func (h *Handler) Login(res http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		slog.Error("Parse form login user: " + err.Error())
		h.ErrMsg(res)
		return
	}

	name := req.FormValue("name")
	password := req.FormValue("password")

	accessToken, err := h.client.Login(req.Context(), name, password)
	if err != nil {
		slog.Error("Create user: " + err.Error())
		return
	}

	urlListEvents := "/list-events/" + accessToken + "?days=0"
	req, err = http.NewRequest(http.MethodGet, urlListEvents, nil)
	if err != nil {
		slog.Error("Create new request: " + err.Error())
		return
	}

	http.Redirect(res, req, urlListEvents, http.StatusFound)
}
