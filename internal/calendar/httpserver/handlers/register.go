package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/storage"
	httpresponse "github.com/mrvin/calendar/pkg/http/response"
	"golang.org/x/crypto/bcrypt"
)

type UserCreator interface {
	CreateUser(ctx context.Context, user *storage.User) error
}

type RequestRegister struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Password string `json:"password" validate:"required,min=6,max=32"`
	Email    string `json:"email"    validate:"required,email"`
}

func NewRegister(creator UserCreator) HandlerFunc {
	validate := validator.New()
	return func(res http.ResponseWriter, req *http.Request) (context.Context, int, error) {
		ctx := req.Context()

		// Read json request
		var request RequestRegister
		body, err := io.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			return ctx, http.StatusBadRequest, fmt.Errorf("read body request: %w", err)
		}
		if err := json.Unmarshal(body, &request); err != nil {
			return ctx, http.StatusBadRequest, fmt.Errorf("unmarshal body request: %w", err)
		}
		ctx = logger.WithUsername(ctx, request.Username)

		// Validation
		if err := validate.Struct(request); err != nil {
			var vErrors validator.ValidationErrors
			if errors.As(err, &vErrors) {
				return ctx, http.StatusBadRequest, fmt.Errorf("invalid request: tag: %s value: %s", vErrors[0].Tag(), vErrors[0].Value())
			}
			return ctx, http.StatusInternalServerError, fmt.Errorf("validation: %w", err)
		}

		hashPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			return ctx, http.StatusInternalServerError, fmt.Errorf("generate hash password: %w", err)
		}

		user := storage.User{
			Name:         request.Username,
			HashPassword: string(hashPassword),
			Email:        request.Email,
			Role:         "user",
		}

		if err = creator.CreateUser(ctx, &user); err != nil {
			err = fmt.Errorf("saving user to storage: %w", err)
			if errors.Is(err, storage.ErrUserExists) {
				return ctx, http.StatusConflict, err
			}
			return ctx, http.StatusInternalServerError, err
		}

		// Write json response
		httpresponse.WriteOK(res, http.StatusCreated)

		return ctx, http.StatusCreated, nil
	}
}
