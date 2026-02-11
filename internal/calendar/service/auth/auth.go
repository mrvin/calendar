package authservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	log "github.com/mrvin/hw-otus-go/hw12-15calendar/internal/logger"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

var ErrAuth = errors.New("compare hash and password")

//nolint:tagliatelle
type Conf struct {
	SecretKey           string `yaml:"secret_key"`
	TokenValidityPeriod int    `yaml:"token_validity_period"`
}

type AuthService struct {
	conf   *Conf
	authSt storage.UserStorage
	tr     trace.Tracer // Think about it.
}

func New(st storage.UserStorage, conf *Conf) *AuthService {
	return &AuthService{
		conf,
		st,
		otel.GetTracerProvider().Tracer("Auth service"),
	}
}

func (a *AuthService) CreateUser(ctx context.Context, user *storage.User) error {
	cctx, sp := a.tr.Start(ctx, "CreateUser")
	defer sp.End()

	slog.Debug(
		"Create user",
		slog.String("Name", user.Name),
		slog.String("Hash Password", user.HashPassword),
		slog.String("Email", user.Email),
		slog.String("Role", user.Role),
	)

	return a.authSt.CreateUser(cctx, user)
}

func (a *AuthService) GetUser(ctx context.Context, userName string) (*storage.User, error) {
	cctx, sp := a.tr.Start(ctx, "GetUser")
	defer sp.End()

	return a.authSt.GetUser(cctx, userName)
}

func (a *AuthService) UpdateUser(ctx context.Context, user *storage.User) error {
	cctx, sp := a.tr.Start(ctx, "UpdateUser")
	defer sp.End()

	return a.authSt.UpdateUser(cctx, user)
}

func (a *AuthService) DeleteUser(ctx context.Context, userName string) error {
	cctx, sp := a.tr.Start(ctx, "DeleteUser")
	defer sp.End()

	return a.authSt.DeleteUser(cctx, userName)
}

func (a *AuthService) ListUsers(ctx context.Context) ([]storage.User, error) {
	cctx, sp := a.tr.Start(ctx, "GetAllUsers")
	defer sp.End()

	return a.authSt.ListUsers(cctx)
}

func (a *AuthService) Authenticate(ctx context.Context, username, password string) (string, error) {
	role, err := a.validCredentials(ctx, username, password)
	if err != nil {
		return "", fmt.Errorf("invalid credentials: %w", err)
	}
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"role":     role,
			"iat":      time.Now().Unix(),                                                              // IssuedAt
			"exp":      time.Now().Add(time.Duration(a.conf.TokenValidityPeriod) * time.Minute).Unix(), // ExpiresAt
		},
	)
	tokenString, err := token.SignedString([]byte(a.conf.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to create a token: %w", err)
	}

	return tokenString, nil
}

func (a *AuthService) validCredentials(ctx context.Context, username, password string) (string, error) {
	user, err := a.authSt.GetUser(ctx, username)
	if err != nil {
		return "", fmt.Errorf("get user by name: %w", err)
	}

	slog.Debug(
		"Compare hash and password",
		slog.String("password", password),
		slog.String("hashedPassword", user.HashPassword),
	)
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashPassword), []byte(password)); err != nil {
		return "", fmt.Errorf("%w: %w", ErrAuth, err)
	}

	return user.Role, nil
}
func (a *AuthService) ParseToken(tokenString string) (jwt.MapClaims, error) {
	// Validate token.
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(a.conf.SecretKey), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}

// Authorized is middleware.
func (a *AuthService) Authorized(next http.HandlerFunc) http.HandlerFunc {
	handler := func(res http.ResponseWriter, req *http.Request) {
		authHeaderValue := req.Header.Get("Authorization")
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeaderValue, bearerPrefix) {
			http.Error(res, "request does not contain an Authorization Bearer token", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeaderValue, bearerPrefix)

		claims, err := a.ParseToken(tokenString)
		if err != nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}
		username := claims["username"]

		ctx := log.WithUsername(req.Context(), username.(string))

		next(res, req.WithContext(ctx)) // Pass request to next handler
	}

	return http.HandlerFunc(handler)
}
