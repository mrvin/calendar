package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

//nolint:tagliatelle
type Conf struct {
	SecretKey           string `env:"SECRET_KEY"            yaml:"secret_key"`
	TokenValidityPeriod int    `env:"TOKEN_VALIDITY_PERIOD" yaml:"token_validity_period"` // in minute
}

type Auth struct {
	conf   *Conf
	authSt storage.UserStorage
}

func New(st storage.UserStorage, conf *Conf) *Auth {
	return &Auth{conf, st}
}

func (a *Auth) Login(ctx context.Context, username, password string) (string, error) {
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

func (a *Auth) ParseToken(tokenString string) (jwt.MapClaims, error) {
	// Validate token.
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (any, error) {
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
func (a *Auth) Authorized(next http.HandlerFunc) http.HandlerFunc {
	handler := func(res http.ResponseWriter, req *http.Request) {
		authHeaderValue := req.Header.Get("Authorization")
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeaderValue, bearerPrefix) {
			http.Error(res, "header does not contain an authorization token", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeaderValue, bearerPrefix)

		claims, err := a.ParseToken(tokenString)
		if err != nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}
		username := claims["username"]

		ctx := logger.WithUsername(req.Context(), username.(string)) //nolint:forcetypeassert

		next(res, req.WithContext(ctx)) // Pass request to next handler
	}

	return http.HandlerFunc(handler)
}

func GetUsernameFromCtx(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", errors.New("ctx is nil")
	}
	if username, ok := ctx.Value(logger.ContextKeyUsername).(string); ok {
		return username, nil
	}

	return "", errors.New("no username in ctx")
}

func (a *Auth) validCredentials(ctx context.Context, username, password string) (string, error) {
	user, err := a.authSt.GetUser(ctx, username)
	if err != nil {
		return "", fmt.Errorf("get user by name: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashPassword), []byte(password)); err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidCredentials, err)
	}

	return user.Role, nil
}
