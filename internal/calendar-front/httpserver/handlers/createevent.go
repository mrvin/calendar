package handlers

import (
	"log/slog"
	"net/http"
	"time"
)

func (h *Handler) DisplayFormEvent(res http.ResponseWriter, req *http.Request) {
	accessToken := req.PathValue("accessToken")
	data := struct {
		Title string
		Body  struct {
			AccessToken string
		}
	}{
		Title: "Create event",
		Body: struct {
			AccessToken string
		}{accessToken},
	}
	if err := h.templates.Execute("form-event.html", res, data); err != nil {
		slog.Error("Execute display form event template: " + err.Error())
		h.ErrMsg(res)
		return
	}
}

func (h *Handler) CreateEvent(res http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		slog.Error("Parse create event form: " + err.Error())
		h.ErrMsg(res)
		return
	}
	accessToken := req.PathValue("accessToken")

	title := req.FormValue("title")
	description := req.FormValue("description")
	startTime := req.FormValue("startTime")
	stopTime := req.FormValue("stopTime")
	timeZone := req.FormValue("timezone")

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		slog.Error("Get location: " + err.Error())
		h.ErrMsg(res)
		return
	}

	timeNow := time.Now().In(loc)
	tZone := timeNow.Format("Z07:00")
	startTimeGO, err := time.ParseInLocation(time.RFC3339, startTime+":00"+tZone, loc)
	if err != nil {
		slog.Error("Parse start time location: " + err.Error())
		h.ErrMsg(res)
		return
	}

	stopTimeGO, err := time.ParseInLocation(time.RFC3339, stopTime+":00"+tZone, loc)
	if err != nil {
		slog.Error("Parse stop time location: " + err.Error())
		h.ErrMsg(res)
		return
	}
	if _, err := h.client.CreateEvent(req.Context(), accessToken, title, description, startTimeGO, stopTimeGO); err != nil {
		slog.Error("Create event: " + err.Error())
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
