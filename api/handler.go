package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/SquareX-Labs/adk/errors"
	"github.com/SquareX-Labs/adk/respond"
)

// API Handler's ---------------------------------------------------------------

// Handler custom api handler help us to handle all the errors in one place
type Handler func(w http.ResponseWriter, r *http.Request) *errors.AppError

// ServeHTTP implements http handler interface
func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// NOTE: this expects the hijacked responsewriter to catch the statuscode
	// that will be logged in logger middleware.
	// use logger middleware at the root of the router chain as possible

	url := fmt.Sprintf("%s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
	// hijack the request body and enclose again
	body, err := hijackBody(r)
	if err != nil {
		notifySlack(url, body, err)
	}

	if err := fn(w, r); err != nil {
		notifySlack(url, body, err)
		// TODO: handle 5XX, notify developers. Configurable
		if fErr := respond.Fail(w, err); fErr.NotNil() {
			log.Printf("[panic] failed to write response. [%s] %s [%d] %s?%s",
				getReqID(r), r.Method, err.GetStatus(), r.URL.Path, r.URL.RawQuery,
			)
		}
	}
}

// RequestIDHeader is the name of the HTTP Header which contains the request id.
// Exported so that it can be changed by developers
var RequestIDHeader = "X-Request-Id"

func getReqID(r *http.Request) string {
	return r.Context().Value(RequestIDHeader).(string)
}

// notifySlack notifies the slack channel with the error message
func notifySlack(url string, body []byte, err *errors.AppError) {
	if slackHookInfo == "" {
		return
	}
	errType := "client-err"
	if err.GetStatus() > 499 {
		errType = "server-err"
	}

	prefix := fmt.Sprintf("%s ATTENTION %s %d", ServiceName, errType, err.GetStatus())
	message := err.Error()

	errStr := fmt.Sprintf("%s: \n```Path: %s\n Payload: %s\n\n Msg: %s```", prefix, url, string(body), message)
	data := map[string]string{"text": errStr}

	jsonData, _ := json.Marshal(data)

	http.Post(slackHookInfo, "application/json", bytes.NewBuffer(jsonData))
}

var hijackMethods = map[string]bool{
	http.MethodPost: true,
	http.MethodPut:  true,
}

func hijackBody(r *http.Request) ([]byte, *errors.AppError) {
	var (
		body []byte
		err  error
	)

	if !hijackMethods[r.Method] {
		return []byte("nil"), nil
	}
	if r.Body == nil {
		return []byte("nil"), nil
	}
	body, err = io.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("[ADK Notifier] HijackRequestBody: read body %s\n", err.Error())
		return []byte("nil"), errors.InternalServer(msg)
	}

	if err := r.Body.Close(); err != nil {
		msg := fmt.Sprintf("[ADK Notifier] HijackRequestBody: close body, %s", err.Error())
		return []byte("nil"), errors.InternalServer(msg)
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}
