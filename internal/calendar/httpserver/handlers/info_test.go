package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestInfo(t *testing.T) {
	handler := ErrorHandler("Info", Info)
	t.Run("Success smoke test", func(t *testing.T) {
		t.Parallel()

		tag = "v1.2.3"
		hash = "993415b3e201086e12c89c2b3f5ad2153aed2f5d"
		date = time.Date(2025, time.November, 29, 10, 0, 0, 0, time.UTC).String()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/info", nil)
		if err != nil {
			t.Fatalf("create new request: %v", err)
		}

		handler.ServeHTTP(res, req)

		status := http.StatusOK
		if res.Code != status {
			t.Errorf("expected status code %d but received %d", status, res.Code)
		}
		var response ResponseInfo
		json.Unmarshal(res.Body.Bytes(), &response)
		if response.Tag != tag {
			t.Errorf(`expected tag "%s" but received "%s"`, tag, response.Tag)
		}
		if response.Hash != hash {
			t.Errorf(`expected hash "%s" but received "%s"`, hash, response.Hash)
		}
		if response.Date != date {
			t.Errorf(`expected date "%s" but received "%s"`, date, response.Date)
		}
	})
}
