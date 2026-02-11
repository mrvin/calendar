package signin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	authenticate "github.com/mrvin/hw-otus-go/hw12-15calendar/internal/calendar/service/auth"
	log "github.com/mrvin/hw-otus-go/hw12-15calendar/internal/logger"
	httpresponse "github.com/mrvin/hw-otus-go/hw12-15calendar/pkg/http/response"
)

type UserAuth interface {
	Authenticate(ctx context.Context, username, password string) (string, error)
}

type Request struct {
	Username string `example:"Bob"    json:"username" validate:"required,min=3,max=20"`
	Password string `example:"qwerty" json:"password" validate:"required,min=6,max=32"`
}

type Response struct {
	Token  string `example:"abJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ8.eyJleHAiOjE3MDQ3MzExODUsImlhdCI6MTcwNDczMDI4NSwidXNlcm5hbWUiOiJCb2IifQ.XDV9U8Wu202vp5g0gJFma7t5oVZXZlAhN-TMPBOZqEA" json:"token"`
	Status string `example:"OK"                                                                                                                                                      json:"status"`
}

// @Router		/signin [get].
func New(auth UserAuth) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		op := "Signin: "
		ctx := req.Context()

		// Read json request
		var request Request
		body, err := io.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			err := fmt.Errorf("read body request: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(body, &request); err != nil {
			err := fmt.Errorf("unmarshal body request: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = log.WithUsername(ctx, request.Username)

		slog.DebugContext(
			ctx,
			"Signin",
			slog.String("password", request.Password),
		)

		// Validation
		if err := validator.New().Struct(request); err != nil {
			var vErr *validator.ValidationErrors
			if errors.As(err, &vErr) {
				err := fmt.Errorf("invalid request: %w", vErr)
				slog.ErrorContext(ctx, op+err.Error())
				httpresponse.WriteError(res, err.Error(), http.StatusBadRequest)
				return
			}
			err := fmt.Errorf("ValidationErrors not found in the error chain: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := auth.Authenticate(ctx, request.Username, request.Password)
		if err != nil {
			err := fmt.Errorf("authentication: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			if errors.Is(err, authenticate.ErrAuth) {
				httpresponse.WriteError(res, err.Error(), http.StatusUnauthorized)
			} else {
				httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Write json response
		response := Response{
			Token:  token,
			Status: "OK",
		}
		jsonResponse, err := json.Marshal(&response)
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

		slog.InfoContext(ctx, "Signin")
	}
}
