package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/storage"
)

type EventsLister interface {
	ListEvents(ctx context.Context, username string, start, end time.Time) ([]storage.Event, error)
}

type ResponseListEvents struct {
	Events []storage.Event `json:"events"`
	Status string          `json:"status"`
}

func NewListEvents(lister EventsLister) HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) (context.Context, int, error) {
		ctx := req.Context()

		username, err := auth.GetUsernameFromCtx(ctx)
		if err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("getting username from ctx: %w", err)
		}
		ctx = logger.WithUsername(ctx, username)

		startStr := req.URL.Query().Get("start_time")
		endStr := req.URL.Query().Get("end_time")

		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return ctx, http.StatusBadRequest, errors.New("invalid start_time format, use RFC3339")
		}
		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return ctx, http.StatusBadRequest, errors.New("invalid end_time format, use RFC3339")
		}
		if start.After(end) {
			return ctx, http.StatusBadRequest, errors.New("start_time must be before or equal to end_time")
		}

		events, err := lister.ListEvents(ctx, username, start, end)
		if err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("getting list events from storage: %w", err)
		}

		response := ResponseListEvents{
			Events: events,
			Status: "OK",
		}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("marshal response: %w", err)
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		if _, err := res.Write(jsonResponse); err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("write response: %w", err)
		}

		return ctx, http.StatusOK, nil
	}
}
