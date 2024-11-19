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
	// Роут для обработки информации об архиве
	s.router.HandleFunc("/api/archive/information", handler.HandleArchiveInformation).Methods("POST")
	s.router.HandleFunc("/api/archive/createArhive", handler.HandleCreateArchive).Methods("POST")
	s.router.HandleFunc("/api/mail/file", handler.HandleFileAndEmails).Methods("POST")
}
