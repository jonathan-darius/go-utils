package response

import (
	"fmt"
	"net/http"
)

// custom error messages
const (
	CounterExceeded = `You have reached the limit for today. `
)

// Response default response messages for http status coes
var Response = []interface{}{
	http.StatusOK:                  "Success",
	http.StatusCreated:             "Successfully Inserted Data",
	http.StatusNoContent:           "",
	http.StatusBadRequest:          "Bad Request",
	http.StatusUnauthorized:        "Unauthorized",
	http.StatusForbidden:           "Forbidden",
	http.StatusNotFound:            "Not Found",
	http.StatusUnprocessableEntity: "Unprocessable Entity",
	http.StatusInternalServerError: "Internal Server Error",
	http.StatusNotImplemented:      "Not Implemented",
	http.StatusGone:                "Resource Expired",
	http.StatusConflict:            "Resource Conflict",
}

// PleaseWait generates message for given duration
func PleaseWait(msg string, timePassed, duration int) string {
	timeEstimation := msg + "Please wait for"
	if timePassed < 60 {
		timeEstimation += fmt.Sprintf(" %d second(s).", 60-timePassed)
		return timeEstimation
	}

	estimatedHour := (duration - timePassed) / 3600
	estimatedMinute := ((duration - timePassed) - (estimatedHour * 3600)) / 60

	if estimatedHour > 0 {
		timeEstimation += fmt.Sprintf(" %d hour(s)", estimatedHour)
	}
	timeEstimation += fmt.Sprintf(" %d minute(s).", estimatedMinute)

	return timeEstimation
}
