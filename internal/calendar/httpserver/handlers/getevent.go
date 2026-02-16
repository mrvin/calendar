package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/storage"
)

type EventGetter interface {
	GetEvent(ctx context.Context, username string, id uuid.UUID) (*storage.Event, error)
}

//nolint:tagliatelle
type ResponseGetEvent struct {
	ID           uuid.UUID      `json:"id"`
	Title        string         `json:"title"`
	Description  string         `json:"description,omitempty"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	NotifyBefore *time.Duration `json:"notify_before,omitempty"`
	Status       string         `json:"status"`
}

func NewGetEvent(getter EventGetter) HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) (context.Context, int, error) {
		ctx := req.Context()

		idStr := req.PathValue("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return ctx, http.StatusBadRequest, fmt.Errorf("parse id: %w", err)
		}

		username, err := auth.GetUsernameFromCtx(ctx)
		if err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("getting username from ctx: %w", err)
		}
		ctx = logger.WithUsername(ctx, username)

		event, err := getter.GetEvent(ctx, username, id)
		if err != nil {
			err := fmt.Errorf("getting event from storage: %w", err)
			if errors.Is(err, storage.ErrEventNotFound) {
				return ctx, http.StatusNotFound, err
			}
			return ctx, http.StatusInternalServerError, err
		}

		// Write json response
		response := ResponseGetEvent{
			ID:           event.ID,
			Title:        event.Title,
			Description:  event.Description,
			StartTime:    event.StartTime,
			EndTime:      event.EndTime,
			NotifyBefore: event.NotifyBefore,
			Status:       "OK",
		}
		jsonResponseEvent, err := json.Marshal(response)
		if err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("marshal response: %w", err)
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		if _, err := res.Write(jsonResponseEvent); err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("write response: %w", err)
		}

		return ctx, http.StatusOK, nil
	}
}
