package handlers

import (
	"context"
	"log/slog"
	"net/http"

	httpresponse "github.com/mrvin/calendar/pkg/http/response"
)

type HandlerFunc func(http.ResponseWriter, *http.Request) (context.Context, int, error)

func ErrorHandler(msg string, handler HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, code, err := handler(res, req)
		if err != nil {
			slog.ErrorContext(ctx, msg, slog.String("error", err.Error())) //nolint:contextcheck
			httpresponse.WriteError(res, err.Error(), code)
			return
		}
		slog.DebugContext(ctx, msg) //nolint:contextcheck
	}
}
