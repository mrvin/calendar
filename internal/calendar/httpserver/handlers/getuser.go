package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	log "github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/storage"
	httpresponse "github.com/mrvin/calendar/pkg/http/response"
)

type UserGetter interface {
	GetUser(ctx context.Context, name string) (*storage.User, error)
}

type ResponseGetUser struct {
	Name         string `example:"Bob"            json:"name"`
	HashPassword string `example:"$3a$10$8.nlQZMRbgpjNNpZzQnZ4OPsuUo1HQ/XFe93qc2tPjBEYlMVFe43W" json:"hashPassword"`
	Email        string `example:"email@mail.com" json:"email"`
	Role         string `example:"user"           json:"role"`
	Status       string `example:"OK"             json:"status"`
}

func NewGetUser(getter UserGetter) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		op := "Get user info: "
		ctx := req.Context()

		username, err := log.GetUsernameFromCtx(ctx)
		if err != nil {
			err := fmt.Errorf("get username from ctx: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			return
		}

		user, err := getter.GetUser(ctx, username)
		if err != nil {
			err := fmt.Errorf("get user info from storage: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			if errors.Is(err, storage.ErrNoUser) {
				httpresponse.WriteError(res, err.Error(), http.StatusNotFound)
			} else {
				httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Write json response
		response := ResponseGetUser{
			Name:         user.Name,
			HashPassword: user.HashPassword,
			Email:        user.Email,
			Role:         user.Role,
			Status:       "OK",
		}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			err := fmt.Errorf("marshal response: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		if _, err := res.Write(jsonResponse); err != nil {
			err := fmt.Errorf("write response: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			return
		}

		slog.InfoContext(ctx, "Get user info")
	}
}
