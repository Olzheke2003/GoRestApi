package apiserver

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type APIserver struct {
	config *Config
	logger *logrus.Logger
	router *mux.Router
}

func (s *APIserver) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)

	return nil
}
func (s *APIserver) Start() error {
	fmt.Println("Server is starting...") // Для отладки

	if err := s.configureLogger(); err != nil {
		return err
	}

	s.configureRouter()
	fmt.Println("qwe")
	s.logger.Info("starting api server")
	err := http.ListenAndServe(s.config.BindAddr, s.router)
	if err != nil {
		fmt.Println("Error starting server:", err) // Печать ошибки, если сервер не запускается
	}
	return err
}
