package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/storage"
)

type EventCreator interface {
	CreateEvent(ctx context.Context, event *storage.Event) (uuid.UUID, error)
}

//nolint:tagliatelle
type RequestCreateEvent struct {
	Title        string         `json:"title"                   validate:"required,min=2,max=64"`
	Description  string         `json:"description,omitempty"   validate:"omitempty,min=2,max=512"`
	StartTime    time.Time      `json:"start_time"              validate:"required"`
	EndTime      time.Time      `json:"end_time"                validate:"required"`
	NotifyBefore *time.Duration `json:"notify_before,omitempty" validate:"omitempty"`
}

type ResponseCreateEvent struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func NewCreateEvent(creator EventCreator) HandlerFunc {
	validate := validator.New()
	return func(res http.ResponseWriter, req *http.Request) (context.Context, int, error) {
		ctx := req.Context()

		username, err := auth.GetUsernameFromCtx(ctx)
		if err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("getting username from ctx: %w", err)
		}
		ctx = logger.WithUsername(ctx, username)

		// Read json request
		var request RequestCreateEvent
		body, err := io.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			return ctx, http.StatusBadRequest, fmt.Errorf("read body request: %w", err)
		}
		if err := json.Unmarshal(body, &request); err != nil {
			return ctx, http.StatusBadRequest, fmt.Errorf("unmarshal body request: %w", err)
		}

		// Validation
		if err := validate.Struct(request); err != nil {
			var vErrors validator.ValidationErrors
			if errors.As(err, &vErrors) {
				return ctx, http.StatusBadRequest, fmt.Errorf("invalid request: tag: %s value: %s", vErrors[0].Tag(), vErrors[0].Value())
			}
			return ctx, http.StatusInternalServerError, fmt.Errorf("validation: %w", err)
		}
		if request.StartTime.After(request.EndTime) {
			return ctx, http.StatusBadRequest, errors.New("start_time must be before or equal to end_time")
		}

		event := storage.Event{
			Title:        request.Title,
			Description:  request.Description,
			StartTime:    request.StartTime,
			EndTime:      request.EndTime,
			NotifyBefore: request.NotifyBefore,
			Username:     username,
		}
		id, err := creator.CreateEvent(ctx, &event)
		if err != nil {
			err = fmt.Errorf("saving event to storage: %w", err)
			if errors.Is(err, storage.ErrDateBusy) {
				return ctx, http.StatusConflict, err
			}
			return ctx, http.StatusInternalServerError, err
		}

		// Write json response
		response := ResponseCreateEvent{
			ID:     id,
			Status: "OK",
		}
		jsonResponse, err := json.Marshal(&response)
		if err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("marshal response: %w", err)
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		if _, err := res.Write(jsonResponse); err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("write response: %w", err)
		}

		return ctx, http.StatusOK, nil
	}
}
