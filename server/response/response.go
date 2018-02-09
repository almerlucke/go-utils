// Package response defines a default JSON response format with success flag, payload
// and errors. Convenience response methods are provided.
package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorSection is a section for specific errors
type ErrorSection string

// ErrorReasons is a slice of error reasons for a section
type ErrorReasons []string

// ErrorMap for an error response. With a map multiple error sections can be
// returned to the user. For each error section (key) an array of strings can be
// given
type ErrorMap map[ErrorSection]ErrorReasons

// Error error interface
func (errs ErrorMap) Error() string {
	desc := ""

	for key, value := range errs {
		desc += fmt.Sprintf("%v: %v\n", key, value)
	}

	return desc
}

// Reason creates an error map with a generic reason section with one
// error reason
func Reason(str string) ErrorMap {
	return ErrorMap{
		"reason": ErrorReasons{str},
	}
}

// Response structure to be returned as json for each json route
type Response struct {
	Success bool        `json:"success"`
	Payload interface{} `json:"payload,omitempty"`
	Errors  ErrorMap    `json:"errors,omitempty"`
}

// Write a response
func (r *Response) Write(rw http.ResponseWriter, statusCode int) {
	js, err := json.Marshal(r)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	rw.Write(js)
}

/*
	Response convenience methods
*/

// InternalServerError writes an internal server error with a reason
func InternalServerError(rw http.ResponseWriter, reason string) {
	r := &Response{
		Success: false,
		Payload: nil,
		Errors:  Reason(reason),
	}

	r.Write(rw, http.StatusInternalServerError)
}

// ValidationError writes a (possible) validation error. If error is of type
// ErrorMap a bad request is written, otherwise an internal server error
func ValidationError(rw http.ResponseWriter, err error) {
	if errorMap, ok := err.(ErrorMap); ok {
		BadRequest(rw, errorMap)
	} else {
		InternalServerError(rw, err.Error())
	}
}

// Unauthorized writes an unauthorized response with a reason
func Unauthorized(rw http.ResponseWriter, reason string) {
	r := &Response{
		Success: false,
		Payload: nil,
		Errors:  Reason(reason),
	}

	r.Write(rw, http.StatusUnauthorized)
}

// Forbidden writes a forbidden response with a reason
func Forbidden(rw http.ResponseWriter, reason string) {
	r := &Response{
		Success: false,
		Payload: nil,
		Errors:  Reason(reason),
	}

	r.Write(rw, http.StatusForbidden)
}

// Accepted writes an accepted response
func Accepted(rw http.ResponseWriter, payload interface{}) {
	r := &Response{
		Success: true,
		Payload: payload,
		Errors:  nil,
	}

	r.Write(rw, http.StatusAccepted)
}

// Created writes a created response
func Created(rw http.ResponseWriter, payload interface{}) {
	r := &Response{
		Success: true,
		Payload: payload,
		Errors:  nil,
	}

	r.Write(rw, http.StatusCreated)
}

// OK writes a successful response
func OK(rw http.ResponseWriter, payload interface{}) {
	r := &Response{
		Success: true,
		Payload: payload,
		Errors:  nil,
	}

	r.Write(rw, http.StatusOK)
}

// BadRequest writes a bad request
func BadRequest(rw http.ResponseWriter, errs ErrorMap) {
	r := &Response{
		Success: false,
		Payload: nil,
		Errors:  errs,
	}

	r.Write(rw, http.StatusBadRequest)
}

// NotFound writes a not found request
func NotFound(rw http.ResponseWriter) {
	r := &Response{
		Success: false,
		Payload: nil,
		Errors:  Reason("404 page not found"),
	}

	r.Write(rw, http.StatusNotFound)
}

// MethodNotAllowed writes a method not allowed response
func MethodNotAllowed(rw http.ResponseWriter) {
	r := &Response{
		Success: false,
		Payload: nil,
		Errors:  Reason("405 method not allowed"),
	}

	r.Write(rw, http.StatusMethodNotAllowed)
}
