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
	ListEvents(ctx context.Context, username string, startWindow, endWindow time.Time) ([]storage.Event, error)
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

		dateStr := req.URL.Query().Get("date")            // date - дата в формате YYYY-MM-DD.
		weekStartStr := req.URL.Query().Get("week_start") // week_start - дата начала недели в формате YYYY-MM-DD.
		monthStr := req.URL.Query().Get("month")          // month - месяц в формате YYYY-MM.
		countParams := uint8(0)
		if dateStr != "" {
			countParams++
		}
		if weekStartStr != "" {
			countParams++
		}
		if monthStr != "" {
			countParams++
		}
		if countParams != 1 {
			return ctx, http.StatusBadRequest, errors.New("exactly one of date, week_start, month must be provided")
		}
		//TODO: validet
		var startWindow, endWindow time.Time
		switch {
		case dateStr != "":
			startWindow, err = time.Parse(time.DateOnly, dateStr)
			if err != nil {
				return ctx, http.StatusBadRequest, errors.New("invalid date format, use YYYY-MM-DD")
			}
			endWindow = startWindow.AddDate(0, 0, 1)
		case weekStartStr != "":
			startWindow, err = time.Parse(time.DateOnly, weekStartStr)
			if err != nil {
				return ctx, http.StatusBadRequest, errors.New("invalid week_start format, use YYYY-MM-DD")
			}
			endWindow = startWindow.AddDate(0, 0, 7)

		case monthStr != "":
			startWindow, err = time.Parse("2025-12", monthStr)
			if err != nil {
				return ctx, http.StatusBadRequest, errors.New("invalid month format, use YYYY-MM")
			}
			endWindow = startWindow.AddDate(0, 1, 0)
		}

		events, err := lister.ListEvents(ctx, username, startWindow, endWindow)
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
