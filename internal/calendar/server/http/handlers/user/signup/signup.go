package signup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	log "github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/storage"
	httpresponse "github.com/mrvin/calendar/pkg/http/response"
	"golang.org/x/crypto/bcrypt"
)

type UserCreator interface {
	CreateUser(ctx context.Context, user *storage.User) error
}

type Request struct {
	Username string `example:"Bob"            json:"username" validate:"required,min=3,max=20"`
	Password string `example:"qwerty"         json:"password" validate:"required,min=6,max=32"`
	Email    string `example:"email@mail.com" json:"email"    validate:"required,email"`
}

// @Router		/signup [post].
func New(creator UserCreator) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		op := "Signup: "
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
			"Signup",
			slog.String("password", request.Password),
			slog.String("email", request.Email),
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

		hashPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			err := fmt.Errorf("generate hash password: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			return
		}

		user := storage.User{
			Name:         request.Username,
			HashPassword: string(hashPassword),
			Email:        request.Email,
			Role:         "user",
		}

		if err = creator.CreateUser(ctx, &user); err != nil {
			err := fmt.Errorf("saving user to storage: %w", err)
			slog.ErrorContext(ctx, op+err.Error())
			httpresponse.WriteError(res, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write json response
		httpresponse.WriteOK(res, http.StatusCreated)

		slog.InfoContext(ctx, "Signup")
	}
}
