package utils

import (
	"encoding/json"
	"flare-indexer/services/api"
	"fmt"
	"net/http"
)

// Decode body from the request into value.
// Any error is written into the response and false is returned.
// (It is enough to just return from the request handler on false value)
//
// TODO: add validation from go-playground/validator
func DecodeBody(w http.ResponseWriter, r *http.Request, value any) bool {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&value)
	if err != nil {
		http.Error(w, fmt.Sprint("error parsing request body: %w", err), http.StatusBadRequest)
		return false
	}
	return true
}

// Write value into w as json. Handles possible error as internal server error
func WriteResponse(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(&value)
	if err != nil {
		http.Error(w, fmt.Sprint("error writing reponse: %w", err), http.StatusInternalServerError)
	}
}

// Writes value as data field in ApiResponse
// Handles error as internal server error
func WriteApiResponse(w http.ResponseWriter, status api.ApiResponseStatus, value any) {
	response := api.ApiResponse{
		Status: status,
		Data:   value,
	}
	WriteResponse(w, response)
}

// Equivalent to WriteApiResponse with status ApiResponseStatusOk
func WriteApiResponseOk(w http.ResponseWriter, value any) {
	WriteApiResponse(w, api.ApiResponseStatusOk, value)
}

// Set InternalServerError to output if err is not nil. Return true if err is not nil
func HandleInternalServerError(w http.ResponseWriter, err error) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}
