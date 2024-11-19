package apiserver

import (
	handler "github.com/Olzheke2003/GoRestApi/internal/app/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func New(config *Config) *APIserver {
	return &APIserver{
		config: config,
		logger: logrus.New(),
		router: mux.NewRouter(),
	}
}

func (s *APIserver) configureRouter() {
	s.router.HandleFunc("/api/archive/information", handler.HandleArchiveInformation).Methods("POST")
	s.router.HandleFunc("/api/archive/files", handler.HandleCreateArchive).Methods("POST")
	s.router.HandleFunc("/hello", handler.Hello).Methods("GET")
}
