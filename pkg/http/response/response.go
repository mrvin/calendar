package response

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type OK struct {
	Status string `json:"status"`
}

type Error struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func WriteOK(res http.ResponseWriter, status int) {
	response := OK{
		Status: "OK",
	}
	jsonResponse, err := json.Marshal(&response)
	if err != nil {
		err := fmt.Errorf("WriteOK: marshal response: %w", err)
		slog.Error(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(status)
	if _, err := res.Write(jsonResponse); err != nil {
		err := fmt.Errorf("WriteOK: write response: %w", err)
		slog.Error(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

func WriteError(res http.ResponseWriter, description string, status int) {
	response := Error{
		Status: "Error",
		Error:  description,
	}

	jsonResponse, err := json.Marshal(&response)
	if err != nil {
		err := fmt.Errorf("WriteError: marshal error: %w", err)
		slog.Error(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(status)
	if _, err := res.Write(jsonResponse); err != nil {
		err := fmt.Errorf("WriteError: write error: %w", err)
		slog.Error(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}
