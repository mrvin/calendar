package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

//nolint:gochecknoglobals
var (
	tag  string
	hash string
	date string
)

type ResponseInfo struct {
	Tag  string `json:"tag"`
	Hash string `json:"hash"`
	Date string `json:"date"`
}

func Info(res http.ResponseWriter, req *http.Request) (context.Context, int, error) {
	ctx := req.Context()

	response := ResponseInfo{
		Tag:  tag,
		Hash: hash,
		Date: date,
	}
	jsonResponse, err := json.Marshal(&response)
	if err != nil {
		return ctx, http.StatusInternalServerError, fmt.Errorf("marshal response: %w", err)
	}
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := res.Write(jsonResponse); err != nil {
		return ctx, http.StatusInternalServerError, fmt.Errorf("write response: %w", err)
	}

	return ctx, http.StatusOK, nil
}
