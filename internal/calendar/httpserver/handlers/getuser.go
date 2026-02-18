package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/storage"
)

type UserGetter interface {
	GetUser(ctx context.Context, username string) (*storage.User, error)
}

type ResponseGetUser struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

func NewGetUser(getter UserGetter) HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) (context.Context, int, error) {
		ctx := req.Context()

		username, err := auth.GetUsernameFromCtx(ctx)
		if err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("getting username from ctx: %w", err)
		}

		user, err := getter.GetUser(ctx, username)
		if err != nil {
			err := fmt.Errorf("getting user from storage: %w", err)
			if errors.Is(err, storage.ErrUserNotFound) {
				return ctx, http.StatusNotFound, err
			}
			return ctx, http.StatusInternalServerError, err
		}

		// Write json response
		response := ResponseGetUser{
			Name:   user.Name,
			Email:  user.Email,
			Role:   user.Role,
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
