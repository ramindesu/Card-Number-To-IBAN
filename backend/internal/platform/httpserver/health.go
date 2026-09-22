package httpserver

import (
	"io"
	"net/http"
)

// Health reports that the API process is running.
func Health(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(writer, "{\"status\":\"ok\"}\n")
}
