package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	authentication "github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/logger"
)

type UserLoginer interface {
	Login(ctx context.Context, username, password string) (string, error)
}

type RequestLogin struct {
	Username string `json:"username"`
	Password string `json:"password"` //nolint:gosec
}

type ResponseLogin struct {
	Token  string `json:"token"`
	Status string `json:"status"`
}

func NewLogin(auth UserLoginer) HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) (context.Context, int, error) {
		ctx := req.Context()

		// Read json request
		var request RequestLogin
		body, err := io.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			return ctx, http.StatusBadRequest, fmt.Errorf("read body request: %w", err)
		}
		if err := json.Unmarshal(body, &request); err != nil {
			return ctx, http.StatusBadRequest, fmt.Errorf("unmarshal body request: %w", err)
		}
		ctx = logger.WithUsername(ctx, request.Username)

		token, err := auth.Login(ctx, request.Username, request.Password)
		if err != nil {
			err = fmt.Errorf("logining: %w", err)
			if errors.Is(err, authentication.ErrInvalidCredentials) {
				return ctx, http.StatusUnauthorized, err
			}
			return ctx, http.StatusInternalServerError, err
		}

		// Write json response
		response := ResponseLogin{
			Token:  token,
			Status: "OK",
		}
		jsonResponse, err := json.Marshal(&response)
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
