package get

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
	httpresponse "github.com/mrvin/hw-otus-go/hw12-15calendar/pkg/http/response"
)

type EventGetter interface {
	GetEvent(ctx context.Context, id int64) (*storage.Event, error)
}

//nolint:tagliatelle
type Response struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	StartTime   time.Time `json:"start_time"`
	StopTime    time.Time `json:"stop_time,omitempty"`
	UserName    string    `json:"user_name"`
	Status      string    `json:"status"`
}

func New(getter EventGetter) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		op := "Get event: "
		ctx := req.Context()
		idStr := req.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			err := fmt.Errorf("convert id: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusBadRequest)
			return
		}

		event, err := getter.GetEvent(ctx, id)
		if err != nil {
			err := fmt.Errorf("get event from storage: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			if errors.Is(err, storage.ErrNoEvent) {
				httpresponse.WriteError(res, err.Error(), http.StatusBadRequest)
			} else {
				httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Write json response
		response := Response{
			ID:          event.ID,
			Title:       event.Title,
			Description: event.Description,
			StartTime:   event.StartTime,
			StopTime:    event.StopTime,
			UserName:    event.UserName,
			Status:      "OK",
		}
		jsonResponseEvent, err := json.Marshal(response)
		if err != nil {
			err := fmt.Errorf("marshal response: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		if _, err := res.Write(jsonResponseEvent); err != nil {
			err := fmt.Errorf("write response: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			return
		}

		slog.InfoContext(ctx, "Get event")
	}
}
