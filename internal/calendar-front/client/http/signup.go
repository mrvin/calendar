package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RequestSignUp struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type ResponseSignUp struct {
	ID     int64  `json:"id,omitempty"`
	Status string `json:"status,required"`
	Error  string `json:"error,omitempty"`
}

func (c *Client) SignUp(ctx context.Context, name, password, email string) (int64, error) {
	request := RequestSignUp{
		UserName: name,
		Password: password,
		Email:    email,
	}

	jsonResponse, err := json.Marshal(&request)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: marshal request: %w", err)
	}
	reqHTTP, err := http.NewRequestWithContext(ctx, "POST", "https://%s:%d/signup", bytes.NewReader(jsonResponse))
	resp, err := c.Do(reqHTTP)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: do https requste: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: read all: %w", err)
	}

	var response ResponseSignUp
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("CreateUser: unmarshal body response: %w", err)
	}
	if response.Status != "OK" {
		return 0, fmt.Errorf("CreateUser: %s", response.Error)
	}

	return response.ID, nil
}
