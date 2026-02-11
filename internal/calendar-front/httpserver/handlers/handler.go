package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/calendar-front/client"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/calendar-front/httpserver/templateloader"
)

var ErrIDEmpty = errors.New("id is empty")

type Handler struct {
	templates *templateloader.TemplateLoader
	client    client.Calendar
}

func New(client client.Calendar) *Handler {
	templates := templateloader.New()
	templates.Load("templates")
	return &Handler{
		templates: templates,
		client:    client,
	}
}

type resp struct {
	Title string
	Body  struct {
		Text string
	}
}

func getID(req *http.Request) (int64, error) {
	idStr := req.URL.Query().Get("id")
	if idStr == "" {
		return 0, ErrIDEmpty
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("convert id: %w", err)
	}

	return id, nil
}

func (h *Handler) ErrMsg(res http.ResponseWriter) {
	text := resp{Title: "Error", Body: struct{ Text string }{"Sorry something went wrong."}}
	if err := h.templates.Execute("text.html", res, text); err != nil {
		slog.Error("Execut error tetemplate: " + err.Error())
		return
	}
}
