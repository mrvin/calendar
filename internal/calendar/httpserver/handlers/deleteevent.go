package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/storage"
	httpresponse "github.com/mrvin/calendar/pkg/http/response"
)

type EventDeleter interface {
	DeleteEvent(ctx context.Context, username string, id uuid.UUID) error
}

func NewDeleteEvent(deleter EventDeleter) HandlerFunc {
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

		if err := deleter.DeleteEvent(ctx, username, id); err != nil {
			err := fmt.Errorf("deleting event from storage: %w", err)
			if errors.Is(err, storage.ErrEventNotFound) {
				return ctx, http.StatusNotFound, err
			}
			return ctx, http.StatusInternalServerError, err
		}

		// Write json response
		httpresponse.WriteOK(res, http.StatusNoContent)

		return ctx, http.StatusNoContent, nil
	}
}
