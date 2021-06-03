package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

// httpResponse writes handler json into http.ResponseWriter
func httpResponse(w http.ResponseWriter, c int, b []byte) {
	w.WriteHeader(c)
	_, _ = w.Write(b)
}

// httpErrorResponse parses handler message into ApiError struct
//   writes handler json into http.ResponseWriter
func HttpErrorResponse(w http.ResponseWriter, c int, m string) {
	b, _ := json.Marshal(ApiError{Message: m})
	log.Printf("\033[1;31m%v: %s\033[0m", c, m)
	httpResponse(w, c, b)
}

type ApiError struct {
	Message string `json:"message"`
}
