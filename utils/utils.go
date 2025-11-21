package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gengeo7/workmate/logger"
	"github.com/gengeo7/workmate/repository"
	"github.com/gengeo7/workmate/types"
)

type Response struct {
	Data   any
	Status int
}

type ApiError struct {
	StatusCode    int
	OriginalError error
	Msg           string
}

func (a ApiError) Error() string {
	return a.Msg
}

func NewApiError(code int, msg string, err error) *ApiError {
	return &ApiError{
		StatusCode:    code,
		Msg:           msg,
		OriginalError: err,
	}
}

func SendError(w http.ResponseWriter, r *http.Request, err error) {
	var response types.ErrorRes
	var statusCode int
	var ae *ApiError
	if errors.As(err, &ae) {
		statusCode = ae.StatusCode
		response.Error = ae.Msg
		if r != nil && ae.OriginalError != nil {
			logger.Error("internal error",
				"route", r.RequestURI,
				"error", ae.OriginalError.Error(),
			)
		}

	} else {
		statusCode = http.StatusInternalServerError
		response.Error = "unhandled internal error"
		if r != nil {
			logger.Error("unhandled internal error",
				"route", r.RequestURI,
				"error", err.Error(),
			)
		} else {
			logger.Error("unhandled internal error", "error", err.Error())
		}
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func SendResponse(response *Response, err error, w http.ResponseWriter, r *http.Request) {
	if err != nil {
		SendError(w, r, err)
		return
	}

	if response == nil {
		response = &Response{
			Data: types.MessageRes{
				Message: "ok",
			},
			Status: http.StatusOK,
		}
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(response.Status)
	if err := json.NewEncoder(w).Encode(response.Data); err != nil {
		SendError(w, r, err)
	}
}

type ErrorCreator = func(err error) *ApiError

type ErrDbCase struct {
	Func     func(error) bool
	Creator  ErrorCreator
	CheckErr bool
}

func DeadlineDbError(err error) *ApiError {
	return NewApiError(http.StatusRequestTimeout, "достигнут лимит по времени", err)
}

func UnhandledError(err error) *ApiError {
	return NewApiError(http.StatusInternalServerError, "непредвиденная ошибка", err)
}

func NotFound(err error) *ApiError {
	return NewApiError(http.StatusNotFound, "не удалось найти по id", err)
}

func TestDbErr(err error, cases ...*ErrDbCase) error {
	for _, c := range cases {
		if c.Func(err) {
			if c.CheckErr {
				return c.Creator(err)
			} else {
				return c.Creator(nil)
			}
		}
	}
	if repository.IsErrDeadline(err) {
		return DeadlineDbError(err)
	}

	return UnhandledError(err)
}
