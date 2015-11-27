package web

import (
    "encoding/json"
    "net/http"

    "github.com/GitbookIO/micro-analytics/web/errors"
)

func render(w http.ResponseWriter, data interface{}, err error) {
    // Error handling
    if err != nil {
        renderError(w, err)
        return
    }

    // Check content
    if data == nil {
        return
    }

    // Finally write JSON
    w.Header().Set("Content-Type", "application/json")
    jsonMarshal(w, data)
}

func renderError(w http.ResponseWriter, err error) {
    // Cast error to a RequestError
    if rqError, ok := err.(*errors.RequestError); ok {
        renderRequestError(w, rqError)
        return
    }

    defaultErr := errors.InternalError
    defaultErr.Message = err.Error()
    renderRequestError(w, &defaultErr)
}

func renderRequestError(w http.ResponseWriter, err *errors.RequestError) {
    w.WriteHeader(err.StatusCode())

    w.Header().Set("Content-Type", "application/json")
    if err := jsonMarshal(w, err); err != nil {
        genericError(w, err)
    }
}

func genericError(w http.ResponseWriter, err error) {
    http.Error(w, err.Error(), 500)
}

func jsonMarshal(w http.ResponseWriter, data interface{}) error {
    return json.NewEncoder(w).Encode(data)
}