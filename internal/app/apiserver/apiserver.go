package apiserver

import (
	"fmt"
	"net/http"
)

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
