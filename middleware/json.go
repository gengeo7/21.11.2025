package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gengeo7/workmate/utils"
)

type ValidateJsonKey struct{}

func ValidateJson[T any]() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := r.Body
			defer body.Close()

			var v T
			decoder := json.NewDecoder(body)
			decoder.DisallowUnknownFields()
			err := decoder.Decode(&v)
			if err == io.EOF {
				utils.SendError(w, r, utils.NewApiError(http.StatusBadRequest, "пустое тело запроса", nil))
				return
			}
			if err != nil {
				utils.SendError(w, r, utils.NewApiError(http.StatusBadRequest, "ошибка чтения тела запроса", nil))
				return
			}

			ctx := context.WithValue(r.Context(), ValidateJsonKey{}, v)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func DtoFromContext[T any](ctx context.Context) *T {
	val := ctx.Value(ValidateJsonKey{})
	if res, ok := val.(T); !ok {
		return nil
	} else {
		return &res
	}
}
