package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/getaddrinfo/proxy-fingerprint-scraper/database"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

var codeNoFingerprints = "no_fingerprints"

func (s *Server) HandleGetRandomFingerprint(w http.ResponseWriter, r *http.Request) {
	fp, err := s.db.GetRandomFingerprint()
	w.Header().Set("Content-Type", "application/json")

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		WriteError(w, "No Fingerprints", &codeNoFingerprints, http.StatusNotFound)
		return
	}

	if err != nil {
		WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(GetFingerprintResponse{
		ID:          fp.ID,
		Fingerprint: fp.Fingerprint,
		ProxyIP:     fp.ProxyIP,
	})

	if err != nil {
		WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) HandleGetSpecificFingerprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	value, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		WriteError(w, "Bad Request: id is not a valid uint64", nil, http.StatusBadRequest)
		return
	}

	fp, err := s.db.GetSpecificFingerprint(value)
	w.Header().Set("Content-Type", "application/json")

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		WriteError(w, "Not Found", nil, http.StatusNotFound)
		return
	}

	if err != nil {
		WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(GetFingerprintResponse{
		ID:          fp.ID,
		Fingerprint: fp.Fingerprint,
		ProxyIP:     fp.ProxyIP,
	})

	if err != nil {
		WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) HandleGetAllFingerprintsJson(w http.ResponseWriter, r *http.Request) {
	fps, err := s.db.GetAllFingerprints()
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		zap.S().Named("api.get_all_fingerprints").Error(err.Error())
		WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
		return
	}

	if fps == nil {
		fps = []string{}
	}

	data, err := json.Marshal(fps)

	if err != nil {
		WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func (s *Server) HandleGetAllFingerprintsRaw(w http.ResponseWriter, r *http.Request) {
	fps, err := s.db.GetAllFingerprints()

	if err != nil {
		zap.S().Named("api.get_all_fingerprints").Error(err.Error())
		WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
		return
	}

	out := ""

	for _, fp := range fps {
		out += fmt.Sprintf("%s\n", fp)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(out))
}

func (s *Server) HandleGetMeta(w http.ResponseWriter, r *http.Request) {
	result, err := s.db.CountFingerprints()

	if err != nil {
		WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
		return
	}

	user := r.Context().Value("user").(database.GetAuthResult)

	txt, err := RenderHomeTemplate(result, user, s.pm)

	if err != nil {
		zap.L().Named("api.home").Error(err.Error())
		WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
		return
	}

	w.Write([]byte(txt))
}

func (s *Server) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	result, err := s.db.GetAllUsers()

	if err != nil {
		WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
		return
	}

	txt, err := RenderAdminTemplate(result)

	if err != nil {
		zap.L().Named("api.home").Error(err.Error())
		WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
		return
	}

	w.Write([]byte(txt))
}
