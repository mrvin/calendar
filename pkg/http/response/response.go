package response

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type ResponsetOK struct {
	Status string `example:"OK" json:"status"`
}

type ResponseError struct {
	Status string `example:"Error"             json:"status"`
	Error  string `example:"error description" json:"error"`
}

func WriteOK(res http.ResponseWriter, status int) {
	response := ResponsetOK{
		Status: "OK",
	}
	jsonResponse, err := json.Marshal(&response)
	if err != nil {
		err := fmt.Errorf("WriteOK: marshal error: %w", err)
		slog.Error(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	if _, err := res.Write(jsonResponse); err != nil {
		err := fmt.Errorf("WriteOK: write error: %w", err)
		slog.Error(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

func WriteError(res http.ResponseWriter, msgError string, status int) {
	response := ResponseError{
		Status: "Error",
		Error:  msgError,
	}

	jsonResponse, err := json.Marshal(&response)
	if err != nil {
		err := fmt.Errorf("WriteError: marshal error: %w", err)
		slog.Error(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	if _, err := res.Write(jsonResponse); err != nil {
		err := fmt.Errorf("WriteError: write error: %w", err)
		slog.Error(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}
