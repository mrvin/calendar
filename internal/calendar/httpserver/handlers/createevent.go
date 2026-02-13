package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/storage"
	httpresponse "github.com/mrvin/calendar/pkg/http/response"
)

type EventCreator interface {
	CreateEvent(ctx context.Context, event *storage.Event) (int64, error)
}

//nolint:tagliatelle
type RequestCreateEvent struct {
	Title       string    `json:"title"       validate:"required,min=2,max=64"`
	Description string    `json:"description" validate:"omitempty,min=2,max=512"`
	StartTime   time.Time `json:"start_time"  validate:"required"`
	StopTime    time.Time `json:"stop_time"   validate:"required"`
}

type ResponseCreateEvent struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

func NewCreateEvent(creator EventCreator) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userName, err := auth.GetUsernameFromCtx(req.Context())
		if userName == "" {
			err := fmt.Errorf("CreateEvent: %w", err)
			slog.Error(err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusBadRequest)
			return
		}

		// Read json request
		var request RequestCreateEvent

		body, err := io.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			err := fmt.Errorf("CreateEvent: read body req: %w", err)
			slog.Error(err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &request); err != nil {
			err := fmt.Errorf("CreateEvent: unmarshal body request: %w", err)
			slog.Error(err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusBadRequest)
			return
		}

		slog.Debug(
			"Create event",
			slog.String("title", request.Title),
			slog.String("description", request.Description),
			slog.Time("start time", request.StartTime),
			slog.Time("stop time", request.StopTime),
		)

		// Validation
		if err := validator.New().Struct(request); err != nil {
			errors := err.(validator.ValidationErrors)
			err := fmt.Errorf("CreateEvent: invalid request: %s", errors)
			slog.Error(err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusBadRequest)
			return
		}

		event := storage.Event{
			Title:       request.Title,
			Description: request.Description,
			StartTime:   request.StartTime,
			StopTime:    request.StopTime,
			UserName:    userName,
		}

		id, err := creator.CreateEvent(req.Context(), &event)
		if err != nil {
			err := fmt.Errorf("CreateEvent: saving event to storage: %w", err)
			slog.Error(err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write json response
		response := ResponseCreateEvent{
			ID:     id,
			Status: "OK",
		}

		jsonResponse, err := json.Marshal(&response)
		if err != nil {
			err := fmt.Errorf("CreateEvent: marshal response: %w", err)
			slog.Error(err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		if _, err := res.Write(jsonResponse); err != nil {
			err := fmt.Errorf("CreateEvent: write response: %w", err)
			slog.Error(err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			return
		}

		slog.Info("New event was created successfully")
	}
}
