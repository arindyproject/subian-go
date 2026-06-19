package httputil

import "net/http"

func NotFoundStatus(err error, notFoundMsg string) int {
	if err.Error() == notFoundMsg {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}
