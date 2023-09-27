package api

import (
	"fmt"
	"net/http"

	"github.com/getaddrinfo/proxy-fingerprint-scraper/common"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/database"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/proxy"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Server struct {
	router *mux.Router
	db     *database.Database
	pm     proxy.Manager
	port   int
}

func NewServer(db *database.Database, pm proxy.Manager, port int) *Server {
	return &Server{
		router: mux.NewRouter(),
		db:     db,
		pm:     pm,
		port:   port,
	}
}

func (s *Server) Run() {
	zap.S().Info("server starting at http://localhost:", s.port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.router)

	if err != nil {
		zap.S().Named("api").Error(err.Error())
	}
}

func (s *Server) InitRoutes() {
	// viewable pages
	s.router.Handle("/", s.AuthMiddleware(http.HandlerFunc(s.HandleGetMeta), common.PermissionViewHomePage))
	s.router.Handle("/admin", s.AuthMiddleware(http.HandlerFunc(s.HandleAdmin), common.PermissionAdmin))

	// api
	s.router.Handle("/api/fingerprints", s.AuthMiddleware(http.HandlerFunc(s.HandleGetAllFingerprintsJson), common.PermissionUseAPI))
	s.router.Handle("/api/fingerprints/raw", s.AuthMiddleware(http.HandlerFunc(s.HandleGetAllFingerprintsRaw), common.PermissionUseAPI))
	s.router.Handle("/api/fingerprints/random", s.AuthMiddleware(http.HandlerFunc(s.HandleGetRandomFingerprint), common.PermissionUseAPI))
	s.router.Handle("/api/fingerprints/{id:[0-9]+}", s.AuthMiddleware(http.HandlerFunc(s.HandleGetSpecificFingerprint), common.PermissionAdmin))
}
