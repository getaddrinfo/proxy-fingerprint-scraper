package api

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func WriteError(w http.ResponseWriter, msg string, code *string, status int) {
	data, err := json.Marshal(ErrorResponse{
		Error: msg,
		Code:  code,
	})

	if err != nil {

		data, err := json.Marshal(ErrorResponse{
			Error: "Internal Server Error",
		})

		if err != nil {
			zap.S().Named("api.error.write").Error(err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}

		w.WriteHeader(status)
		w.Write(data)

		return
	}

	w.WriteHeader(status)
	w.Write(data)
}

func WriteBadRequest(w http.ResponseWriter) {
	WriteError(w, "Bad Request", nil, http.StatusBadRequest)
}
