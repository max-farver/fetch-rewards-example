package server

import (
	"context"
	"database/sql"
	"fetch-rewards/common"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	Ctx    context.Context
	DB     *sql.DB
	Router *mux.Router
	Logger *zap.SugaredLogger
	Validator *validator.Validate

	PointsService common.PointsService
}

func (s *Server) InitServer() {
	s.Router = mux.NewRouter().StrictSlash(true)
	s.registerRouteHandlers()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Printf("defaulting to port %s", port)
	}

	withMiddleware := cors.
		Default().
		Handler(s.Router)

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, withMiddleware))
}

func (s *Server) registerRouteHandlers() {
	s.Router.HandleFunc("/health", healthCheck).Methods("GET")
	s.RegisterTransactionRoutes()
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")

	_, _ = w.Write([]byte("Healthy"))
}