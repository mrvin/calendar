package delete

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/mrvin/calendar/internal/storage"
	httpresponse "github.com/mrvin/calendar/pkg/http/response"
)

type EventDeleter interface {
	DeleteEvent(ctx context.Context, id int64) error
}

func New(deleter EventDeleter) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		op := "Delete event: "
		ctx := req.Context()
		idStr := req.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			err := fmt.Errorf("convert id: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err := deleter.DeleteEvent(ctx, id); err != nil {
			err := fmt.Errorf("delete event from storage: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			if errors.Is(err, storage.ErrNoEvent) {
				httpresponse.WriteError(res, err.Error(), http.StatusBadRequest)
			} else {
				httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Write json response
		httpresponse.WriteOK(res, http.StatusOK)

		slog.Info("Delete event")
	}
}
