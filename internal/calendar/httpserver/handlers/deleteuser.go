package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/storage"
	httpresponse "github.com/mrvin/calendar/pkg/http/response"
)

type UserDeleter interface {
	DeleteUser(ctx context.Context, username string) error
}

func NewDeleteUser(deleter UserDeleter) HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) (context.Context, int, error) {
		ctx := req.Context()

		username, err := auth.GetUsernameFromCtx(ctx)
		if err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("getting username from ctx: %w", err)
		}
		ctx = logger.WithUsername(ctx, username)

		if err := deleter.DeleteUser(ctx, username); err != nil {
			err = fmt.Errorf("deleting user from storage: %w", err)
			if errors.Is(err, storage.ErrUserNotFound) {
				return ctx, http.StatusNotFound, err
			}
			return ctx, http.StatusInternalServerError, err
		}

		// Write json response
		httpresponse.WriteOK(res, http.StatusNoContent)

		return ctx, http.StatusNoContent, nil
	}
}
