package handler

import (
	"context"
	"errors"
)

var ErrUserNameIsEmpty = errors.New("user name is empty")

func GetUserNameFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if userName, ok := ctx.Value("username").(string); ok {
		return userName
	}

	return ""
}
