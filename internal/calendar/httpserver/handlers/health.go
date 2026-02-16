package handlers

import (
	"net/http"

	httpresponse "github.com/mrvin/calendar/pkg/http/response"
)

func Health(res http.ResponseWriter, _ *http.Request) {
	httpresponse.WriteOK(res, http.StatusOK)
}
